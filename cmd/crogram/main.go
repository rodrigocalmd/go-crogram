// Command crogram encodes and decodes text with substitution ciphers.
//
// Usage:
//
//	crogram encode [flags] [text]
//	crogram decode [flags] [text]
//	crogram key    [flags]
//	crogram puzzle [flags] [text]
//
// The text to work on is given as arguments, or on standard input when it is
// omitted or "-". Encoding prints the key to standard error, so the coded text
// stays pipeable:
//
//	echo "meet me at midnight" | crogram encode -seed 42
//	crogram decode -key "$KEY" "coded text"
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/rodrigocalmd/go-crogram"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch command := os.Args[1]; command {
	case "encode", "decode", "key", "puzzle":
		if err := run(command, os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "crogram:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "crogram: unknown command %q\n\n", command)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `crogram encodes and decodes text with substitution ciphers.

Usage:
  crogram encode [flags] [text]   encode text, printing the key to stderr
  crogram decode [flags] [text]   decode text with a key or a seed
  crogram key    [flags]          print a key for the chosen cipher
  crogram puzzle [flags] [text]   encode text and keep the answer

Text is taken from the arguments, or from standard input when it is omitted
or given as "-". Flags come before the text, and a text that itself starts
with a dash has to be separated by "--":

  crogram decode -seed 42 "coded text"     flags first
  crogram encode -- "-dash-led text"       "--" for text that looks like a flag
  echo "piped text" | crogram encode       text on standard input

decode needs one of -key, -seed or -passphrase; encode, key and puzzle work
without one.

Flags:
  -key string         use a key produced by "crogram key"
  -seed int           derive a reproducible cipher from a seed
  -passphrase string  derive a reproducible cipher from a passphrase
  -charset string     use this character set instead of the default
  -derangement        forbid a rune standing in for itself (default true)

Examples:
  echo "meet me at midnight" | crogram encode -seed 42
  crogram decode -key "$KEY" "coded text"
  crogram key -passphrase "correct horse battery staple"
`)
}

// run executes one subcommand.
func run(command string, args []string) error {
	fs := flag.NewFlagSet("crogram "+command, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	key := fs.String("key", "", "key produced by \"crogram key\"")
	seed := fs.Int64("seed", 0, "seed to derive the cipher from")
	passphrase := fs.String("passphrase", "", "passphrase to derive the cipher from")
	charset := fs.String("charset", "", "character set to use instead of the default")
	derangement := fs.Bool("derangement", true, "forbid a rune standing in for itself")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	// Remember which flags the caller actually set, so that -seed 0 still
	// means "seed 0" rather than "no seed".
	var given []string
	fs.Visit(func(f *flag.Flag) { given = append(given, f.Name) })

	// A flag written after the text is the mistake worth catching here: the
	// flag package stops at the first positional argument, so a trailing
	// "-seed 42" becomes part of the text — silently, on encode.
	if late, ok := lateFlag(args, fs.Args()); ok {
		return fmt.Errorf("%s was given after the text, and flags have to come before it:\n"+
			"    %s\n"+
			"    run crogram help for the flags", late, synopsis(command))
	}

	origins := 0
	for _, name := range []string{"key", "seed", "passphrase"} {
		if wasGiven(given, name) {
			origins++
		}
	}
	if origins > 1 {
		return errors.New("-key, -seed and -passphrase are mutually exclusive")
	}
	if command == "decode" && origins == 0 {
		return errors.New("decode needs -key, -seed or -passphrase: without one it would decode with a cipher drawn at random")
	}
	if *key != "" && wasGiven(given, "charset") {
		fmt.Fprintln(os.Stderr, "crogram: note: -charset is ignored with -key, which carries its own character set")
	}
	if *key != "" && wasGiven(given, "derangement") {
		fmt.Fprintln(os.Stderr, "crogram: note: -derangement is ignored with -key, which carries its own mapping")
	}

	var options []crogram.Option
	if *charset != "" {
		options = append(options, crogram.WithCharset([]rune(*charset)))
	}
	options = append(options, crogram.WithDerangement(*derangement))

	cipher, err := buildCipher(*key, *seed, *passphrase, given, options)
	if err != nil {
		return err
	}

	if command == "key" {
		if args := fs.Args(); len(args) > 0 {
			return fmt.Errorf("key takes no text argument, got %q", args[0])
		}
		fmt.Println(cipher.Key())
		return nil
	}

	text, err := input(fs.Args())
	if err != nil {
		return err
	}

	switch command {
	case "encode":
		fmt.Println(cipher.Encode(text))
		fmt.Fprintln(os.Stderr, "key:", cipher.Key())
	case "decode":
		fmt.Println(cipher.Decode(text))
	case "puzzle":
		puzzle := crogram.NewPuzzleWith(cipher, text)
		fmt.Println(puzzle)
		fmt.Fprintln(os.Stderr, "key:", puzzle.Key())
	}
	return nil
}

// wasGiven reports whether the caller passed the named flag, which fs.Visit
// tells apart from a flag left at its default value.
func wasGiven(given []string, name string) bool {
	return slices.Contains(given, name)
}

// synopsis is the one-line shape of a subcommand, used to point at the right
// form when an argument lands in the wrong place.
func synopsis(command string) string {
	if command == "key" {
		return "crogram key [flags]"
	}
	return fmt.Sprintf("crogram %s [flags] [text]", command)
}

// lateFlag reports the first argument that looks like a flag but arrived after
// the text. The flag package stops parsing at the first positional argument, so
// "crogram decode text -seed 42" leaves -seed inside the text rather than
// failing — which encode would then cheerfully encode along with the text.
//
// An explicit "--" anywhere in the arguments means the caller separated text
// from flags on purpose, so nothing is reported then.
func lateFlag(args, text []string) (string, bool) {
	if slices.Contains(args, "--") {
		return "", false
	}
	for _, arg := range text {
		if len(arg) > 1 && strings.HasPrefix(arg, "-") {
			return arg, true
		}
	}
	return "", false
}

// buildCipher picks the cipher the caller asked for, in the order of certainty:
// an explicit key beats a passphrase, which beats a seed, which beats a random
// draw.
func buildCipher(key string, seed int64, passphrase string, given []string, options []crogram.Option) (*crogram.Cipher, error) {
	if key != "" {
		return crogram.ParseKey(key)
	}
	if passphrase != "" {
		return crogram.NewFromPassphrase(passphrase, options...), nil
	}
	if wasGiven(given, "seed") {
		return crogram.NewWithSeed(seed, options...), nil
	}
	return crogram.NewCipher(options...), nil
}

// input returns the text to work on: the arguments joined by spaces, or all of
// standard input when no text was given. A trailing newline, which a shell
// always adds, is dropped.
func input(args []string) (string, error) {
	if len(args) > 0 && args[0] != "-" {
		return strings.Join(args, " "), nil
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(raw), "\n"), nil
}

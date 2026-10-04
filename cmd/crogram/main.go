// Command crogram encodes and decodes text with substitution ciphers.
//
// It is intended for entertainment and educational purposes. It is not a
// cryptographically secure encryption method.
//
// Usage:
//
//	crogram [options] encode [text]
//	crogram [options] decode [text]
//
// The text is given as arguments, or streamed from standard input (or -f)
// when omitted or "-". Encode prints the seed to stderr, so decode can use
// it; alternatively use --key to export/import a portable key:
//
//	echo "meet me at midnight" | crogram encode
//	crogram decode -s 13258932 "coded text"
//	crogram key -s 13258932
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/rodrigocalmd/go-crogram"
)

// version is set at build time: -ldflags "-X main.version=v1.2.3".
var version = ""

func versionString() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}

// isTerminal reports whether w is an interactive terminal.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

type options struct {
	seed    int64
	hasSeed bool
	key     string
	charset string
	in, out string
	quiet   bool
	version bool
}

func (o *options) set(fs *flag.FlagSet) {
	seed := func(v string) error {
		var n int64
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil || fmt.Sprint(n) != strings.TrimPrefix(v, "+") {
			return fmt.Errorf("invalid seed %q", v)
		}
		o.seed, o.hasSeed = n, true
		return nil
	}
	fs.Func("s", "seed", seed)
	fs.Func("seed", "seed", seed)
	fs.StringVar(&o.key, "k", o.key, "key")
	fs.StringVar(&o.key, "key", o.key, "key")
	fs.StringVar(&o.charset, "c", o.charset, "charset")
	fs.StringVar(&o.charset, "charset", o.charset, "charset")
	fs.StringVar(&o.in, "f", o.in, "input file")
	fs.StringVar(&o.out, "o", o.out, "output file")
	fs.BoolVar(&o.quiet, "q", false, "quiet")
	fs.BoolVar(&o.quiet, "quiet", false, "quiet")
	fs.BoolVar(&o.version, "version", false, "version")
	fs.BoolVar(&o.version, "v", false, "version")
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var o options

	// parse consumes flags from args and returns the remaining arguments.
	parse := func(name string, args []string) ([]string, int) {
		fs := flag.NewFlagSet(name, flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		o.set(fs)
		if err := fs.Parse(args); err != nil {
			if err == flag.ErrHelp {
				usage(stdout)
				return nil, 0
			}
			fmt.Fprintln(stderr, "crogram:", err)
			return nil, 2
		}
		return fs.Args(), -1
	}

	rest, code := parse("crogram", args)
	if code >= 0 {
		return code
	}
	if o.version {
		fmt.Fprintln(stdout, "crogram", versionString())
		return 0
	}
	if len(rest) == 0 {
		usage(stderr)
		return 2
	}

	command := rest[0]
	switch command {
	case "help":
		usage(stdout)
		return 0
	case "encode", "decode", "key":
	default:
		fmt.Fprintf(stderr, "crogram: unknown command %q (try: encode, decode, key, help)\n", command)
		return 2
	}

	rest, code = parse(command, rest[1:])
	if code >= 0 {
		return code
	}
	if o.version {
		fmt.Fprintln(stdout, "crogram", versionString())
		return 0
	}

	if o.key != "" && o.hasSeed {
		fmt.Fprintln(stderr, "crogram: use either -s/--seed or -k/--key, not both")
		return 2
	}

	if o.key != "" && o.charset != "" {
		fmt.Fprintln(stderr, "crogram: a key already contains its characters; use either -k/--key or -c/--charset, not both")
		return 2
	}

	var c *crogram.Cipher
	switch {
	case o.charset != "":
		cs := resolveCharset(o.charset)
		var seed []int64
		if o.hasSeed {
			seed = []int64{o.seed}
		} else if command == "decode" {
			fmt.Fprintln(stderr, "crogram: decode needs the seed or key used to encode (-s <seed> or -k <key>)")
			return 2
		}
		var err error
		if c, err = crogram.NewWithCharset(cs, seed...); err != nil {
			fmt.Fprintln(stderr, libError(err))
			return 2
		}
	case o.key != "":
		var err error
		if c, err = crogram.FromKey(o.key); err != nil {
			fmt.Fprintln(stderr, libError(err))
			return 2
		}
	case o.hasSeed:
		c = crogram.New(o.seed)
	case command == "decode":
		fmt.Fprintln(stderr, "crogram: decode needs the seed or key used to encode (-s <seed> or -k <key>)")
		return 2
	default:
		c = crogram.New()
	}

	if command == "key" {
		if len(rest) > 0 || o.in != "" || o.out != "" {
			fmt.Fprintln(stderr, "crogram: key only prints the key and takes no text; to encode text use: crogram encode [-c ...] [-s ...] \"text\"")
			return 2
		}
		fmt.Fprintln(stdout, c.Key())
		return 0
	}

	// Encode announces the seed unless the caller already knows the cipher.
	announce := func() {
		if command != "encode" || o.key != "" || o.hasSeed || o.quiet {
			return
		}
		s, _ := c.Seed()
		if isTerminal(stderr) {
			// Friendly hint for people; scripts get the bare number.
			extra := ""
			if o.charset != "" {
				extra = fmt.Sprintf(" -c %q", o.charset)
			}
			fmt.Fprintf(stderr, "seed: %d   (decode with: crogram decode -s %d%s)\n", s, s, extra)
		} else {
			fmt.Fprintln(stderr, s)
		}
	}

	var dst io.Writer = stdout
	if o.out != "" {
		f, err := os.Create(o.out)
		if err != nil {
			fmt.Fprintln(stderr, "crogram:", err)
			return 1
		}
		defer f.Close()
		dst = f
	}
	wrap := c.EncodeWriter
	if command == "decode" {
		wrap = c.DecodeWriter
	}

	// Text given as arguments.
	if len(rest) > 0 && !(len(rest) == 1 && rest[0] == "-") && o.in == "" {
		announce()
		w := wrap(dst)
		fmt.Fprintln(w, strings.Join(rest, " "))
		if err := w.Close(); err != nil {
			fmt.Fprintln(stderr, "crogram:", err)
			return 1
		}
		return 0
	}

	// Stream from a file or stdin.
	src := stdin
	if o.in != "" {
		f, err := os.Open(o.in)
		if err != nil {
			fmt.Fprintln(stderr, "crogram:", err)
			return 1
		}
		defer f.Close()
		src = f
	}
	announce()
	w := wrap(dst)
	if _, err := io.Copy(w, src); err != nil {
		fmt.Fprintln(stderr, "crogram:", err)
		return 1
	}
	if err := w.Close(); err != nil {
		fmt.Fprintln(stderr, "crogram:", err)
		return 1
	}
	return 0
}

// libError formats an error from the library. Its messages already start
// with "crogram:", so the prefix is not repeated.
func libError(err error) string {
	return "crogram: " + strings.TrimPrefix(err.Error(), "crogram: ")
}

// presets are the names accepted by -c instead of a literal list of characters.
var presets = map[string]string{
	"default": crogram.DefaultCharset,
	"pt":      crogram.PortugueseCharset,
	"es":      crogram.SpanishCharset,
	"ru":      crogram.RussianCharset,
}

// resolveCharset returns the preset called v, or v itself as a list of characters.
func resolveCharset(v string) string {
	if p, ok := presets[v]; ok {
		return p
	}
	return v
}

func usage(w io.Writer) {
	fmt.Fprint(w, `encode or decode text with a substitution cipher.

WARNING: for entertainment and educational purposes only. This is NOT a
cryptographically secure encryption method; do not use it to protect
sensitive information.

Usage:
  crogram [options] encode [options] [text]   encode text
  crogram [options] decode [options] [text]   decode text
  crogram [options] key                       print the cipher's key

Options:
  -s, --seed <n>    use a specific seed for reproducibility
  -k, --key <key>   use a key (see "crogram key"); portable across versions
  -c, --charset <x> characters to encode: a preset (default, pt, es, ru) or a
                    list such as "abcáé". Default: a-z, A-Z, 0-9. Needed again
                    on decode (not needed with -k, which carries its own)
  -f <file>         read input from a file instead of arguments/stdin
  -o <file>         write output to a file instead of stdout
  -q, --quiet       do not print the seed on encode
  -v, --version     print the version

The text is read from arguments, or streamed from standard input when none
are given. Encode prints the seed used to standard error (unless -s or -k
is given), so decode can be run with it.

Examples:
  crogram encode "Hello World"
  crogram -s 42 encode "secret message"
  crogram key -s 42
  echo "coded text" | crogram decode -k "$(crogram key -s 42)"
  crogram encode -s 42 -f in.txt -o out.txt
  crogram -c pt -s 42 encode "Ação e emoção"
  crogram -c pt -s 42 decode "..."
`)
}

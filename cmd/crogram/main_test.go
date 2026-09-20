package main

import (
	"strings"
	"testing"
)

// TestDecodeNeedsAnOrigin pins the rule that decode will not silently decode
// with a cipher drawn at random: without a key, a seed or a passphrase the
// output would be meaningless and the caller would never be told.
func TestDecodeNeedsAnOrigin(t *testing.T) {
	err := run("decode", []string{"coded text"})
	if err == nil {
		t.Fatal("decode without -key, -seed or -passphrase should fail")
	}
	if !strings.Contains(err.Error(), "-key") {
		t.Errorf("the error should name the flags to use, got %v", err)
	}

	// The other commands still work without an origin.
	for _, command := range []string{"encode", "key", "puzzle"} {
		if err := run(command, []string{"-seed", "1"}); err != nil {
			t.Errorf("%s -seed 1: %v", command, err)
		}
	}
}

// TestDecodeWithAnOriginWorks checks the paths decode is allowed to take.
func TestDecodeWithAnOriginWorks(t *testing.T) {
	if err := run("decode", []string{"-seed", "1", "coded text"}); err != nil {
		t.Errorf("decode -seed 1: %v", err)
	}
	if err := run("decode", []string{"-passphrase", "correct horse", "coded text"}); err != nil {
		t.Errorf("decode -passphrase: %v", err)
	}
	if err := run("decode", []string{"-key", "crogram1:6162:6261", "ba"}); err != nil {
		t.Errorf("decode -key: %v", err)
	}
}

// TestKeyTakesNoTextArgument checks that the one subcommand with nothing to
// encode says so instead of ignoring what it was handed.
func TestKeyTakesNoTextArgument(t *testing.T) {
	if err := run("key", []string{"stray"}); err == nil {
		t.Error("key with a positional argument should fail")
	}
	if err := run("key", []string{"-seed", "1"}); err != nil {
		t.Errorf("key -seed 1: %v", err)
	}
}

// TestFlagsAfterTheTextAreReported pins the check for the mistake the flag
// package cannot report on its own: it stops parsing at the first positional
// argument, so a trailing flag used to be encoded along with the text.
func TestFlagsAfterTheTextAreReported(t *testing.T) {
	for _, command := range []string{"encode", "decode", "puzzle", "key"} {
		err := run(command, []string{"the text", "-seed", "42"})
		if err == nil {
			t.Errorf("%s with a flag after the text should fail", command)
			continue
		}
		if !strings.Contains(err.Error(), "-seed") || !strings.Contains(err.Error(), "before") {
			t.Errorf("%s: the message should name the flag and the rule, got %v", command, err)
		}
		if !strings.Contains(err.Error(), command) {
			t.Errorf("%s: the message should show that command's own form, got %v", command, err)
		}
	}

	// The same mistake on encode used to exit 0, encoding the flag as text.
	if err := run("encode", []string{"some text", "-seed", "42"}); err == nil {
		t.Error("encode should report a flag after the text rather than encode it")
	}

	// An explicit -- means the text is text, dashes and all.
	if err := run("encode", []string{"--", "-seed", "-42"}); err != nil {
		t.Errorf("text after an explicit -- should be accepted: %v", err)
	}
}

// TestKeyIgnoresCharsetAndDerangementOnlyWithANote checks that the flags a key
// overrides are still accepted, so nothing that used to work stops working.
func TestKeyIgnoresCharsetAndDerangementOnlyWithANote(t *testing.T) {
	const key = "crogram1:6162:6261"

	withFlags := run("key", []string{"-key", key, "-charset", "abcdefgh", "-derangement=false"})
	if withFlags != nil {
		t.Fatalf("key -key with -charset and -derangement: %v", withFlags)
	}
	if err := run("key", []string{"-key", key}); err != nil {
		t.Fatalf("key -key: %v", err)
	}
}

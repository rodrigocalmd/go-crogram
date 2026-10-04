package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func do(stdin string, args ...string) (int, string, string) {
	var o, e bytes.Buffer
	c := run(args, strings.NewReader(stdin), &o, &e)
	return c, o.String(), e.String()
}

func TestRoundTrip(t *testing.T) {
	for _, args := range [][]string{{"-s", "42", "encode", "Hello World"}, {"encode", "-s", "42", "Hello World"}, {"--seed=42", "encode", "Hello World"}} {
		c, out, _ := do("", args...)
		if c != 0 {
			t.Fatalf("encode with args %v: exit code %d, want 0", args, c)
		}
		c, dec, _ := do(out, "decode", "-s", "42")
		if c != 0 || dec != "Hello World\n" {
			t.Fatalf("decode after encode with args %v: exit code %d, output %q, want 0 and \"Hello World\\n\"", args, c, dec)
		}
	}
}

func TestErrors(t *testing.T) {
	for _, args := range [][]string{{}, {"bogus"}, {"decode", "x"}, {"-s", "abc", "encode", "x"}, {"-s"}} {
		if c, _, _ := do("", args...); c != 2 {
			t.Errorf("invalid usage %v: exit code %d, want 2", args, c)
		}
	}
}

func TestKey(t *testing.T) {
	_, key, _ := do("", "key", "-s", "42")
	key = strings.TrimSpace(key)
	_, a, _ := do("", "-s", "42", "encode", "Hello")
	c, b, _ := do("", "-k", key, "encode", "Hello")
	if c != 0 || a != b {
		t.Fatalf("encoding with -k <key of seed 42> gave exit %d and %q, but encoding with -s 42 gave %q; they must be equal", c, b, a)
	}
	if c, out, _ := do(b, "decode", "--key", key); c != 0 || out != "Hello\n" {
		t.Errorf("decode with --key: exit code %d, output %q, want 0 and \"Hello\\n\"", c, out)
	}
	if c, _, _ := do("", "encode", "-k", "bad", "x"); c != 2 {
		t.Errorf("encode with an invalid key: exit code %d, want 2", c)
	}
	if c, _, _ := do("", "encode", "-k", key, "-s", "1", "x"); c != 2 {
		t.Errorf("encode with both -k and -s: exit code %d, want 2", c)
	}
}

func TestFiles(t *testing.T) {
	d := t.TempDir()
	in, out := d+"/in.txt", d+"/out.txt"
	os.WriteFile(in, []byte("Hello World\nline two\n"), 0o600)
	if c, _, _ := do("", "encode", "-s", "9", "-f", in, "-o", out); c != 0 {
		t.Fatalf("encode from file to file: exit code %d, want 0", c)
	}
	if c, got, _ := do("", "decode", "-s", "9", "-f", out); c != 0 || got != "Hello World\nline two\n" {
		t.Errorf("decode from file: exit code %d, output %q, want 0 and the original two lines", c, got)
	}
	if c, _, _ := do("", "encode", "-f", d+"/missing"); c != 1 {
		t.Errorf("encode with a missing input file: exit code %d, want 1", c)
	}
}

func TestVersionQuietHelp(t *testing.T) {
	if c, out, _ := do("", "--version"); c != 0 || !strings.HasPrefix(out, "crogram ") {
		t.Errorf("--version: exit code %d, output %q, want 0 and a line starting with \"crogram \"", c, out)
	}
	if c, _, e := do("", "encode", "-q", "x"); c != 0 || e != "" {
		t.Errorf("encode -q: exit code %d, stderr %q, want 0 and empty stderr", c, e)
	}
	if c, _, e := do("", "encode", "x"); c != 0 || e == "" {
		t.Errorf("encode without -q: exit code %d, stderr %q, want 0 and the seed on stderr", c, e)
	}
	for _, a := range []string{"help", "-h", "--help"} {
		if c, out, _ := do("", a); c != 0 || !strings.Contains(out, "Usage:") {
			t.Errorf("%s: exit code %d, output %q, want 0 and a help text containing \"Usage:\"", a, c, out)
		}
	}
}

func TestCharsetOption(t *testing.T) {
	_, enc, _ := do("", "-c", "pt", "-s", "42", "encode", "Ação e emoção")
	if enc == "Ação e emoção\n" {
		t.Fatal("encoding with -c pt left the text unchanged")
	}
	c, dec, _ := do(enc, "decode", "--charset", "pt", "-s", "42")
	if c != 0 || dec != "Ação e emoção\n" {
		t.Errorf("decode with --charset pt: exit code %d, output %q, want 0 and the original text", c, dec)
	}
	// A literal list of characters works as well.
	_, enc, _ = do("", "-c", "abcáé", "-s", "1", "encode", "cáfé abc")
	if c, dec, _ := do(enc, "decode", "-c", "abcáé", "-s", "1"); c != 0 || dec != "cáfé abc\n" {
		t.Errorf("literal charset: exit code %d, output %q", c, dec)
	}
	// A key carries its own charset.
	_, key, _ := do("", "key", "-c", "ru", "-s", "3")
	_, enc, _ = do("", "-c", "ru", "-s", "3", "encode", "Привет")
	if c, dec, _ := do(enc, "decode", "-k", strings.TrimSpace(key)); c != 0 || dec != "Привет\n" {
		t.Errorf("decode with a key of a custom charset: exit code %d, output %q", c, dec)
	}
}

func TestCharsetOptionErrors(t *testing.T) {
	for _, args := range [][]string{
		{"-c", "a", "-s", "1", "encode", "x"},         // too short
		{"-c", "abca", "-s", "1", "encode", "x"},      // repeated character
		{"-c", "pt", "decode", "x"},                   // decode without seed
		{"-c", "pt", "-k", "whatever", "encode", "x"}, // charset and key together
	} {
		if c, _, _ := do("", args...); c != 2 {
			t.Errorf("invalid usage %v: exit code %d, want 2", args, c)
		}
	}
}

func TestLibraryErrorHasSinglePrefix(t *testing.T) {
	_, _, e := do("", "-c", "a", "-s", "1", "encode", "x")
	if strings.Contains(e, "crogram: crogram:") || !strings.HasPrefix(e, "crogram: ") {
		t.Errorf("error message must start with a single \"crogram: \" prefix, got %q", e)
	}
}

func TestKeyRejectsText(t *testing.T) {
	for _, args := range [][]string{
		{"key", "-s", "1", "some text"},
		{"key", "-s", "1", "-f", "in.txt"},
		{"key", "-s", "1", "-o", "out.txt"},
	} {
		c, out, e := do("", args...)
		if c != 2 || out != "" || !strings.Contains(e, "takes no text") {
			t.Errorf("%v: exit code %d, stdout %q, stderr %q, want 2, empty stdout and a message saying key takes no text", args, c, out, e)
		}
	}
}

// Arbitrary arguments must never panic, and the exit code must be one of the
// documented ones (0 ok, 1 I/O error, 2 invalid usage).
func FuzzCLI(f *testing.F) {
	f.Add("encode", "-s", "42")
	f.Add("-c", "pt", "decode")
	f.Add("key", "-k", "c3:abcbca")
	f.Add("--seed=7", "encode", "texto")
	f.Add("-q", "--version", "x")
	f.Add("decode", "-k", "")
	f.Fuzz(func(t *testing.T, a, b, c string) {
		args := []string{a, b, c}
		// -f and -o touch real files: never let the fuzzer choose paths.
		for _, arg := range args {
			name := strings.TrimLeft(arg, "-")
			if strings.HasPrefix(arg, "-") && (strings.HasPrefix(name, "o") || strings.HasPrefix(name, "f")) {
				t.Skip()
			}
		}
		t.Chdir(t.TempDir())
		code, _, _ := do("some input\n", args...)
		if code < 0 || code > 2 {
			t.Fatalf("args %q: exit code %d, want 0, 1 or 2", args, code)
		}
	})
}

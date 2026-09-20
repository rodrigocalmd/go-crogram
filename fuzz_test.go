package crogram_test

import (
	"testing"

	"github.com/rodrigocalmd/go-crogram"
)

// FuzzRoundTrip checks the property everything rests on, against arbitrary
// input rather than the handful of strings a table test would list. Go's
// fuzzer grows a corpus of interesting inputs — invalid UTF-8, combining
// marks, lone surrogates — which is exactly where a rune-wise substitution can
// go wrong.
//
// Run it explicitly with:
//
//	go test -run '^$' -fuzz FuzzRoundTrip -fuzztime 30s .
func FuzzRoundTrip(f *testing.F) {
	f.Add("Hello World 123!")
	f.Add("")
	f.Add("acção — 42!")
	f.Add("  \t\n  ")
	f.Add("aA0aA0")

	// Invalid UTF-8. A range over the string would replace each offending byte
	// with U+FFFD and grow it to three bytes; the byte-wise walk in substitute
	// exists so that this cannot happen.
	f.Add("\xb7")
	f.Add("a\xb7b")
	f.Add("\xff\xfe")

	c := crogram.NewWithSeed(1)
	f.Fuzz(func(t *testing.T, text string) {
		if got := c.Decode(c.Encode(text)); got != text {
			t.Fatalf("Decode(Encode(%q)) = %q", text, got)
		}
	})
}

// FuzzKeyRoundTrip checks that a key always rebuilds the cipher it came from,
// for every seed rather than a chosen few.
func FuzzKeyRoundTrip(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(-1))

	f.Fuzz(func(t *testing.T, seed int64) {
		c := crogram.NewWithSeed(seed)

		key := c.Key()
		rebuilt, err := crogram.ParseKey(key)
		if err != nil {
			t.Fatalf("ParseKey(%q) for seed %d: %v", key, seed, err)
		}
		if got := rebuilt.Key(); got != key {
			t.Fatalf("the rebuilt key is %q, want %q", got, key)
		}
		if got, want := rebuilt.Encode("Hello World 123!"), c.Encode("Hello World 123!"); got != want {
			t.Fatalf("the rebuilt cipher encodes %q, want %q", got, want)
		}
	})
}

// FuzzParseKeyNeverPanics checks that untrusted text cannot bring the parser
// down, whatever it contains.
func FuzzParseKeyNeverPanics(f *testing.F) {
	f.Add("")
	f.Add("crogram1:6162:6162")
	f.Add("crogram1::")
	f.Add("crogram1:zz:zz")
	f.Add("crogram1:ff:ff")

	f.Fuzz(func(t *testing.T, key string) {
		c, err := crogram.ParseKey(key)
		if err != nil {
			if c != nil {
				t.Fatalf("ParseKey(%q) returned a cipher alongside %v", key, err)
			}
			return
		}
		if c == nil {
			t.Fatalf("ParseKey(%q) returned no cipher and no error", key)
		}
		// A key that parsed must decode what it encodes.
		text := "Hello World 123!"
		if got := c.Decode(c.Encode(text)); got != text {
			t.Fatalf("the parsed cipher does not round-trip: %q", got)
		}
	})
}

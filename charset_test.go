package crogram

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func mustCharset(t *testing.T, cs string, seed int64) *Cipher {
	t.Helper()
	c, err := NewWithCharset(cs, seed)
	if err != nil {
		t.Fatalf("NewWithCharset(%q): %v", cs, err)
	}
	return c
}

func TestDefaultCharsetEqualsNew(t *testing.T) {
	c := mustCharset(t, DefaultCharset, 42)
	in := "consistency matters, Hello World 123"
	if got, want := c.Encode(in), New(42).Encode(in); got != want {
		t.Errorf("NewWithCharset(DefaultCharset, 42) encoded %q, but New(42) encoded %q", got, want)
	}
}

func TestCharsetRoundTrip(t *testing.T) {
	cases := []struct{ name, charset, text string }{
		{"portuguese", PortugueseCharset, "Ação, coração e pão: São João não vai à praça. Ü 123"},
		{"spanish", SpanishCharset, "¿Cómo está el niño? ¡Año nuevo! Ñandú 2025"},
		{"russian", RussianCharset, "Привет, мир! Ёж съел 3 яблока. Ёлка"},
		{"japanese hiragana", "あいうえおかきくけこ", "あさ、こんにちは。かお"},
		{"emoji", "😀😁😂🤣😃😄", "😀 hello 😄 😂"},
		{"mixed ascii and accents", "abcáé", "abc áé xyz"},
		{"mixed scripts in text", PortugueseCharset, "Hello 日本語 Привет — ação"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := mustCharset(t, tc.charset, 7)
			enc := c.Encode(tc.text)
			if enc == tc.text {
				t.Errorf("encoding left %q unchanged", tc.text)
			}
			if dec := c.Decode(enc); dec != tc.text {
				t.Errorf("decoding the encoded text did not give the original: want %q, got %q", tc.text, dec)
			}
			if !utf8.ValidString(enc) {
				t.Errorf("encoded text is not valid UTF-8: %q", enc)
			}
		})
	}
}

func TestCharsetOnlyEncodesItsCharacters(t *testing.T) {
	c := mustCharset(t, PortugueseCharset, 3)
	// Cyrillic, CJK, punctuation and spaces are not in the Portuguese charset.
	in := "Привет 日本 !?, \n\t"
	if got := c.Encode(in); got != in {
		t.Errorf("characters outside the charset must not change: input %q, output %q", in, got)
	}
	// Accented letters ARE encoded, and stay inside the charset.
	got := c.Encode("ç")
	if got == "ç" || !strings.ContainsRune(PortugueseCharset, []rune(got)[0]) {
		t.Errorf("ç should be replaced by another character of the charset, got %q", got)
	}
}

func TestCharsetReproducible(t *testing.T) {
	a, b := mustCharset(t, RussianCharset, 5), mustCharset(t, RussianCharset, 5)
	c := mustCharset(t, RussianCharset, 6)
	in := "Привет, мир"
	if a.Encode(in) != b.Encode(in) {
		t.Error("two ciphers with the same charset and seed must encode the same way")
	}
	if a.Encode(in) == c.Encode(in) {
		t.Error("two ciphers with different seeds encoded the text the same way")
	}
	if s, ok := a.Seed(); s != 5 || !ok {
		t.Errorf("Seed() = (%d, %v), want (5, true)", s, ok)
	}
}

func TestCharsetBijection(t *testing.T) {
	for _, cs := range []string{PortugueseCharset, SpanishCharset, RussianCharset} {
		c := mustCharset(t, cs, 1)
		seen := map[rune]bool{}
		for _, r := range cs {
			e := c.encode.lookup(r)
			if seen[e] {
				t.Fatalf("two different characters are encoded to the same character %q", e)
			}
			seen[e] = true
			if !strings.ContainsRune(cs, e) {
				t.Fatalf("%q was encoded to %q, which is outside the charset", r, e)
			}
			if c.decode.lookup(e) != r {
				t.Fatalf("decoding the encoded %q did not give %q back", r, r)
			}
		}
	}
}

func TestInvalidCharsets(t *testing.T) {
	bad := map[string]string{
		"empty":         "",
		"one character": "a",
		"duplicate":     "abca",
		"invalid UTF-8": "ab\xffcd",
		"replacement":   "ab\uFFFDcd",
	}
	for name, cs := range bad {
		if _, err := NewWithCharset(cs); err == nil {
			t.Errorf("NewWithCharset with a %s charset must return an error, but returned none", name)
		}
	}
}

func TestInvalidUTF8Preserved(t *testing.T) {
	in := "ab\xff\xfecd ação"
	for _, cs := range []string{DefaultCharset, PortugueseCharset} {
		c := mustCharset(t, cs, 1)
		if got := c.Decode(c.Encode(in)); got != in {
			t.Errorf("charset %q: invalid bytes were not preserved: want %q, got %q", cs[:3], in, got)
		}
	}
}

func FuzzCharsetRoundTrip(f *testing.F) {
	f.Add(int64(1), "Ação 123 Привет")
	f.Fuzz(func(t *testing.T, seed int64, s string) {
		c, err := NewWithCharset(PortugueseCharset+"абвгдеёжз", seed)
		if err != nil {
			t.Fatal(err)
		}
		if got := c.Decode(c.Encode(s)); got != s {
			t.Errorf("decoding the encoded text did not give the original: want %q, got %q", s, got)
		}
	})
}

func ExampleNewWithCharset() {
	c, _ := NewWithCharset(PortugueseCharset, 42)
	enc := c.Encode("Ação")
	fmt.Println(c.Decode(enc))
	// Output: Ação
}

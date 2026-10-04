package crogram

import (
	"fmt"
	"testing"
	"unicode/utf8"
)

func TestEncodeDecode(t *testing.T) {
	cipher := New(12345) // Using a fixed seed for a predictable test

	testCases := []struct {
		name string
		text string
	}{
		{"simple lowercase", "helloworld"},
		{"mixed case", "HelloWorld"},
		{"with numbers", "HelloWorld123"},
		{"with punctuation", "Hello, World! 123..."},
		{"empty string", ""},
		{"only numbers", "1234567890"},
		{"only punctuation", "!@#$%^&*()"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := cipher.Encode(tc.text)
			decoded := cipher.Decode(encoded)

			if decoded != tc.text {
				t.Errorf("decoding the encoded text did not give the original: want %q, got %q", tc.text, decoded)
			}
		})
	}
}

func TestReproducibleCipher(t *testing.T) {
	seed := int64(98765)
	text := "This is a test for reproducibility."

	cipher1 := New(seed)
	encoded1 := cipher1.Encode(text)

	cipher2 := New(seed)
	encoded2 := cipher2.Encode(text)

	if encoded1 != encoded2 {
		t.Errorf("two ciphers with the same seed must encode the same way:\nfirst:  %s\nsecond: %s", encoded1, encoded2)
	}

	decoded1 := cipher1.Decode(encoded1)
	if decoded1 != text {
		t.Errorf("decoding the encoded text did not give the original: want %q, got %q", text, decoded1)
	}
}

func TestDifferentSeeds(t *testing.T) {
	seed1 := int64(111)
	seed2 := int64(222)
	text := "This should be encoded differently."

	cipher1 := New(seed1)
	encoded1 := cipher1.Encode(text)

	cipher2 := New(seed2)
	encoded2 := cipher2.Encode(text)

	if encoded1 == encoded2 {
		t.Errorf("two ciphers with different seeds encoded the text the same way: %s", encoded1)
	}

	if encoded1 == text {
		t.Errorf("encoding left the text unchanged; the cipher did not substitute anything")
	}
}

func TestRandomCipherReportsNoExplicitSeed(t *testing.T) {
	text := "a cipher created without a seed"
	c := New()
	if _, seeded := c.Seed(); seeded {
		t.Error("New() without arguments: Seed() said the seed was provided by the caller, but it was drawn at random")
	}
	if got := c.Decode(c.Encode(text)); got != text {
		t.Errorf("decoding the encoded text did not give the original: want %q, got %q", text, got)
	}
}

func TestSeedIsReportedWhenProvided(t *testing.T) {
	if s, ok := New(7).Seed(); s != 7 || !ok {
		t.Errorf("New(7).Seed() = (%d, %v), want (7, true)", s, ok)
	}
}

func TestBijection(t *testing.T) {
	c := New(1)
	seen := map[rune]bool{}
	for _, r := range defaultRunes {
		e := c.encode.lookup(r)
		if seen[e] {
			t.Fatalf("two different characters are encoded to the same character %q", e)
		}
		seen[e] = true
		if c.decode.lookup(e) != r {
			t.Fatalf("decoding the encoded %q did not give %q back", r, r)
		}
	}
}

func TestPassthrough(t *testing.T) {
	c := New(1)
	in := "áé日本 \x00\x7f!?\n😀"
	if got := c.Encode(in); got != in {
		t.Errorf("characters outside a-z, A-Z, 0-9 must not change: input %q, output %q", in, got)
	}
}

func TestKnownOutput(t *testing.T) {
	if got := New(42).Encode("consistency matters"); got != "9ZhB1BQSh9U ktQQS6B" {
		t.Errorf("the output for seed 42 changed (it must stay stable): got %q", got)
	}
}

func FuzzRoundTrip(f *testing.F) {
	f.Add(int64(1), "Hello, World 123")
	f.Fuzz(func(t *testing.T, seed int64, s string) {
		c := New(seed)
		if !utf8.ValidString(s) {
			t.Skip()
		}
		if got := c.Decode(c.Encode(s)); got != s {
			t.Errorf("decoding the encoded text did not give the original: want %q, got %q", s, got)
		}
	})
}

func ExampleNew() {
	c := New(42)
	enc := c.Encode("consistency matters")
	fmt.Println(enc)
	fmt.Println(c.Decode(enc))
	// Output:
	// 9ZhB1BQSh9U ktQQS6B
	// consistency matters
}

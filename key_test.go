package crogram

import (
	"strings"
	"testing"
)

func TestKeyRoundTrip(t *testing.T) {
	c := New(42)
	k := c.Key()
	if len(k) != 62 {
		t.Fatalf("Key() must have 62 characters, got %d", len(k))
	}
	c2, err := FromKey(k)
	if err != nil {
		t.Fatal(err)
	}
	in := "consistency matters, Ünï 123"
	if c.Encode(in) != c2.Encode(in) || c2.Decode(c.Encode(in)) != in {
		t.Error("a cipher rebuilt with FromKey encodes or decodes differently from the original cipher")
	}
	if s, ok := c2.Seed(); s != 0 || ok {
		t.Errorf("FromKey cipher: Seed() = (%d, %v), want (0, false)", s, ok)
	}
	if c2.Key() != k {
		t.Error("Key() of a cipher rebuilt with FromKey is different from the original Key()")
	}
}

func TestFromKeyErrors(t *testing.T) {
	good := New(1).Key()
	bad := map[string]string{
		"short":     good[:61],
		"long":      good + "a",
		"duplicate": "a" + good[1:61] + "a",
		"invalid":   "!" + good[1:],
		"empty":     "",
	}
	for name, k := range bad {
		if _, err := FromKey(k); err == nil {
			t.Errorf("FromKey with a %s key must return an error, but it returned none", name)
		}
	}
}

func TestCharsetKeyRoundTrip(t *testing.T) {
	for _, cs := range []string{PortugueseCharset, RussianCharset, "あいうえお", "abc:def"} {
		c := mustCharset(t, cs, 9)
		k := c.Key()
		c2, err := FromKey(k)
		if err != nil {
			t.Fatalf("FromKey(%q): %v", k, err)
		}
		in := "Ação Привет あい abc:def"
		if c.Encode(in) != c2.Encode(in) || c2.Decode(c.Encode(in)) != in {
			t.Errorf("a cipher rebuilt with FromKey behaves differently (charset %q)", cs)
		}
		if c2.Key() != k {
			t.Errorf("Key() changed after FromKey: %q -> %q", k, c2.Key())
		}
	}
	// The default charset keeps the original 62-character key format.
	if k := New(1).Key(); len(k) != 62 || strings.Contains(k, ":") {
		t.Errorf("the key of the default charset must stay 62 plain characters, got %q", k)
	}
}

func TestCustomKeyErrors(t *testing.T) {
	good := mustCharset(t, "abcde", 1).Key() // c5:abcde<perm>
	bad := map[string]string{
		"count is not a number": "cx:abcdeabcde",
		"count too small":       "c1:aa",
		"wrong length":          good[:len(good)-1],
		"duplicate in charset":  "c3:abaabc",
		"substitute outside":    "c3:abcabz",
		"duplicate substitute":  "c3:abcaab",
		"no colon content":      "c3:",
	}
	for name, k := range bad {
		if _, err := FromKey(k); err == nil {
			t.Errorf("FromKey with %s (%q) must return an error, but returned none", name, k)
		}
	}
}

// FromKey receives arbitrary text, so it must never panic, and every key it
// accepts must describe a working cipher whose Key() can be loaded again.
func FuzzFromKey(f *testing.F) {
	f.Add(New(1).Key())
	f.Add(mustCharsetFuzz(PortugueseCharset, 1).Key())
	f.Add("c3:abcbca")
	f.Add("c3:abc")
	f.Add("c:")
	f.Add("")
	f.Fuzz(func(t *testing.T, key string) {
		c, err := FromKey(key)
		if err != nil {
			return
		}
		const sample = "Hello, Wörld 123 ação \xff"
		if got := c.Decode(c.Encode(sample)); got != sample {
			t.Fatalf("accepted key %q: decoding the encoded text did not give the original: want %q, got %q", key, sample, got)
		}
		again, err := FromKey(c.Key())
		if err != nil {
			t.Fatalf("accepted key %q, but its own Key() %q was rejected: %v", key, c.Key(), err)
		}
		if again.Encode(sample) != c.Encode(sample) {
			t.Fatalf("accepted key %q: the cipher rebuilt from Key() encodes differently", key)
		}
	})
}

func mustCharsetFuzz(charset string, seed int64) *Cipher {
	c, err := NewWithCharset(charset, seed)
	if err != nil {
		panic(err)
	}
	return c
}

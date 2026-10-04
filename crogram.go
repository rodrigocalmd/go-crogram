// Package crogram generates substitution ciphers for letters and digits.
//
// New uses the characters a-z, A-Z and 0-9. NewWithCharset accepts any set
// of characters, so other languages can be encoded too. Characters outside
// the set (spaces, punctuation, emoji...) are never changed.
//
// It is intended for entertainment and education; it is not
// cryptographically secure.
package crogram

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"
)

// tables maps characters to their substitutes. ASCII characters use an array
// (fast); anything else goes through the map, which is nil when the charset
// is ASCII only.
type tables struct {
	ascii [128]rune
	other map[rune]rune
}

func (t *tables) lookup(r rune) rune {
	if r >= 0 && r < rune(len(t.ascii)) {
		return t.ascii[r]
	}
	if v, ok := t.other[r]; ok {
		return v
	}
	return r
}

// Cipher handles the encoding and decoding of text.
type Cipher struct {
	encode, decode tables
	charset        []rune
	// byteWise is true when every character of the charset is ASCII. Then
	// text can be processed byte by byte: bytes >= 0x80 only occur inside
	// multi-byte UTF-8 sequences, which are never substituted.
	byteWise bool
	seed     int64
	seeded   bool
}

// apply maps every ASCII byte of src through t into dst (len(dst) >= len(src);
// dst and src may be the same slice). Only valid when the cipher is byteWise.
func apply(dst, src []byte, t *tables) {
	for i, c := range src {
		if c < utf8.RuneSelf {
			c = byte(t.ascii[c])
		}
		dst[i] = c
	}
}

// appendTranslated appends src translated rune by rune. Bytes that are not
// valid UTF-8 are copied unchanged.
func appendTranslated(dst, src []byte, t *tables) []byte {
	for i := 0; i < len(src); {
		r, size := utf8.DecodeRune(src[i:])
		if r == utf8.RuneError && size == 1 {
			dst = append(dst, src[i])
		} else {
			dst = utf8.AppendRune(dst, t.lookup(r))
		}
		i += size
	}
	return dst
}

func (c *Cipher) translate(text string, t *tables) string {
	if len(text) == 0 {
		return ""
	}
	if c.byteWise {
		out := make([]byte, len(text))
		apply(out, []byte(text), t)
		// out is never modified again, so it can back the string without a copy.
		return unsafe.String(&out[0], len(out))
	}
	var b strings.Builder
	b.Grow(len(text))
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && size == 1 {
			b.WriteByte(text[i])
		} else {
			b.WriteRune(t.lookup(r))
		}
		i += size
	}
	return b.String()
}

// Encode transforms a string based on the Cipher's encoding map.
// Characters not in the charset are returned unchanged.
func (c *Cipher) Encode(text string) string { return c.translate(text, &c.encode) }

// Decode transforms a string based on the Cipher's decoding map.
// Characters not in the charset are returned unchanged.
func (c *Cipher) Decode(text string) string { return c.translate(text, &c.decode) }

// build creates a Cipher where charset[i] is encoded to perm[i].
func build(charset, perm []rune) *Cipher {
	c := &Cipher{charset: charset, byteWise: true}
	for _, t := range []*tables{&c.encode, &c.decode} {
		for i := range t.ascii {
			t.ascii[i] = rune(i)
		}
	}
	set := func(t *tables, from, to rune) {
		if from < rune(len(t.ascii)) {
			t.ascii[from] = to
			return
		}
		if t.other == nil {
			t.other = make(map[rune]rune)
		}
		t.other[from] = to
	}
	for i, orig := range charset {
		set(&c.encode, orig, perm[i])
		set(&c.decode, perm[i], orig)
		if orig >= utf8.RuneSelf || perm[i] >= utf8.RuneSelf {
			c.byteWise = false
		}
	}
	return c
}

// shuffled returns a random permutation of charset.
func shuffled(r *rand.Rand, charset []rune) []rune {
	perm := make([]rune, len(charset))
	copy(perm, charset)
	r.Shuffle(len(perm), func(i, j int) { perm[i], perm[j] = perm[j], perm[i] })
	return perm
}

// randomSeed draws a seed from the OS entropy source, falling back to the
// clock if it is unavailable.
func randomSeed() int64 {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		return time.Now().UnixNano()
	}
	return int64(binary.LittleEndian.Uint64(b[:]))
}

func newSeeded(charset []rune, seed []int64) *Cipher {
	seeded := len(seed) > 0
	var s int64
	if seeded {
		s = seed[0]
	} else {
		s = randomSeed()
	}
	c := build(charset, shuffled(rand.New(rand.NewSource(s)), charset))
	c.seed, c.seeded = s, seeded
	return c
}

// New creates a new Cipher for DefaultCharset (a-z, A-Z, 0-9).
// It can be called with an optional seed to generate a reproducible cipher.
// If no seed is provided, a random cipher is generated.
func New(seed ...int64) *Cipher { return newSeeded(defaultRunes, seed) }

// NewWithCharset is like New but substitutes the characters of charset
// instead of a-z, A-Z and 0-9. The charset is a string of distinct
// characters, at least two; the order matters for reproducibility, since the
// same charset and seed always give the same cipher.
//
// NewWithCharset(DefaultCharset, seed) is identical to New(seed).
func NewWithCharset(charset string, seed ...int64) (*Cipher, error) {
	cs, err := parseCharset(charset)
	if err != nil {
		return nil, err
	}
	return newSeeded(cs, seed), nil
}

// parseCharset validates a charset string.
func parseCharset(s string) ([]rune, error) {
	if !utf8.ValidString(s) {
		return nil, fmt.Errorf("crogram: charset is not valid UTF-8")
	}
	cs := []rune(s)
	if len(cs) < 2 {
		return nil, fmt.Errorf("crogram: charset needs at least 2 characters, got %d", len(cs))
	}
	seen := make(map[rune]bool, len(cs))
	for _, r := range cs {
		if r == utf8.RuneError {
			return nil, fmt.Errorf("crogram: charset must not contain U+FFFD")
		}
		if seen[r] {
			return nil, fmt.Errorf("crogram: charset contains %q more than once", r)
		}
		seen[r] = true
	}
	return cs, nil
}

// Seed returns the seed the cipher was built from, and whether it was
// supplied explicitly or drawn at random.
func (c *Cipher) Seed() (int64, bool) {
	return c.seed, c.seeded
}

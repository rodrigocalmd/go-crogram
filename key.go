package crogram

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Key returns the cipher's substitution as a string. It does not depend on
// the Go version or on the random number generator, so it stays valid
// forever. Pass it to FromKey.
//
// For DefaultCharset the key has 62 characters: the i-th one is what the
// i-th character of "a-zA-Z0-9" is encoded to. For any other charset the key
// is "c<n>:<charset><substitutes>", where n is the number of characters, so
// the key carries its own charset.
func (c *Cipher) Key() string {
	perm := make([]rune, len(c.charset))
	for i, r := range c.charset {
		perm[i] = c.encode.lookup(r)
	}
	if string(c.charset) == DefaultCharset {
		return string(perm)
	}
	return fmt.Sprintf("c%d:%s%s", len(c.charset), string(c.charset), string(perm))
}

// FromKey rebuilds a Cipher from a string returned by Key. It fails if the
// key is malformed or is not a permutation of its charset. The resulting
// Cipher has no seed: Seed reports (0, false).
func FromKey(key string) (*Cipher, error) {
	if !utf8.ValidString(key) {
		return nil, fmt.Errorf("crogram: key is not valid UTF-8")
	}
	if strings.HasPrefix(key, "c") && strings.Contains(key, ":") {
		return fromCustomKey(key)
	}
	return fromPermutation(defaultRunes, []rune(key))
}

func fromCustomKey(key string) (*Cipher, error) {
	i := strings.IndexByte(key, ':')
	n, err := strconv.Atoi(key[1:i])
	if err != nil || n < 2 {
		return nil, fmt.Errorf("crogram: invalid character count %q in key", key[1:i])
	}
	rest := []rune(key[i+1:])
	if len(rest) != 2*n {
		return nil, fmt.Errorf("crogram: key must have %d characters after ':', got %d", 2*n, len(rest))
	}
	cs, err := parseCharset(string(rest[:n]))
	if err != nil {
		return nil, err
	}
	return fromPermutation(cs, rest[n:])
}

// fromPermutation checks that perm is a permutation of charset.
func fromPermutation(charset, perm []rune) (*Cipher, error) {
	if len(perm) != len(charset) {
		return nil, fmt.Errorf("crogram: key must have %d characters, got %d", len(charset), len(perm))
	}
	in := make(map[rune]bool, len(charset))
	for _, r := range charset {
		in[r] = true
	}
	used := make(map[rune]bool, len(perm))
	for _, r := range perm {
		if !in[r] {
			return nil, fmt.Errorf("crogram: invalid character %q in key", r)
		}
		if used[r] {
			return nil, fmt.Errorf("crogram: duplicate character %q in key", r)
		}
		used[r] = true
	}
	return build(charset, perm), nil
}

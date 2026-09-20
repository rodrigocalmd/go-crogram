package crogram

import (
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"slices"
	"unicode/utf8"
)

// builtinCharset is the character set a cipher falls back to when neither the
// caller nor DefaultCharset supplies one. DefaultCharset starts as a copy of it,
// so reassigning DefaultCharset cannot leave the package without a set to fall
// back on.
var builtinCharset = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// DefaultCharset is the character set New uses when none is configured: the 26
// lowercase letters, the 26 uppercase letters and the 10 digits, in that
// order, for 62 runes in total.
//
// It is a package-level variable and a cipher copies it on the way in, so a
// program is free to reassign it before building ciphers:
//
//	crogram.DefaultCharset = []rune("abcdefghijklmnopqrstuvwxyz")
//
// Reassigning it is a global mutation: do it once, at start-up, before any
// goroutine builds a cipher, or two goroutines will race on the variable. An
// empty or nil DefaultCharset falls back to a built-in set, so New still
// produces a usable cipher rather than one whose own key cannot be parsed.
var DefaultCharset = append([]rune(nil), builtinCharset...)

// NewFromPairs builds a cipher from explicit couples — a plaintext rune
// followed by the rune that stands in for it:
//
//	c, err := crogram.NewFromPairs('a', 'q', 'b', 'x', 'c', 'm')
//
// The plaintext runes become the character set and the replacements become the
// alphabet, both in the order given. The two sides are independent: a
// replacement need not belong to the character set, which is what makes
// substitutions such as 'a' → '1' possible.
//
// The replacements must not repeat, or the cipher could not be reversed.
// NewFromPairs reports ErrUnpairedRune for an odd number of runes,
// ErrDuplicateRune for a repeated rune on either side, and nothing else.
// Runes left out of the set are passed through unchanged.
func NewFromPairs(pairs ...rune) (*Cipher, error) {
	if len(pairs) == 0 {
		return nil, fmt.Errorf("%w: no pairs given", ErrUnpairedRune)
	}
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("%w: got %d runes", ErrUnpairedRune, len(pairs))
	}

	charset := make([]rune, 0, len(pairs)/2)
	alphabet := make([]rune, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		charset = append(charset, pairs[i])
		alphabet = append(alphabet, pairs[i+1])
	}

	if err := checkDistinct(charset); err != nil {
		return nil, err
	}
	if err := checkDistinct(alphabet); err != nil {
		return nil, err
	}

	return fromAlphabet(charset, alphabet, 0, false, isDerangement(charset, alphabet), "pairs"), nil
}

// NewFromMapping builds a cipher from a plaintext-to-replacement map:
//
//	c, err := crogram.NewFromMapping(map[rune]rune{'a': 'q', 'b': 'x'})
//
// The character set is the sorted set of map keys, so the result is
// deterministic no matter how the map was built. Replacements must not repeat
// (ErrDuplicateRune) and the map must not be empty (ErrUnpairedRune).
func NewFromMapping(mapping map[rune]rune) (*Cipher, error) {
	if len(mapping) == 0 {
		return nil, fmt.Errorf("%w: empty mapping", ErrUnpairedRune)
	}

	charset := make([]rune, 0, len(mapping))
	for plain := range mapping {
		charset = append(charset, plain)
	}
	slices.Sort(charset)

	alphabet := make([]rune, 0, len(mapping))
	for _, plain := range charset {
		alphabet = append(alphabet, mapping[plain])
	}

	if err := checkDistinct(alphabet); err != nil {
		return nil, err
	}

	return fromAlphabet(charset, alphabet, 0, false, isDerangement(charset, alphabet), "mapping"), nil
}

// sanitize collapses repeated runes in a character set, keeping the first
// occurrence of each, and falls back to a sanitized builtinCharset when nothing
// is left — including when DefaultCharset itself was emptied, so the result is
// never empty. It always returns a fresh slice.
func sanitize(charset []rune) []rune {
	out := dedupe(charset)
	if len(out) == 0 {
		out = dedupe(builtinCharset)
	}
	return out
}

// dedupe keeps the first occurrence of each rune, in order.
func dedupe(runes []rune) []rune {
	out := make([]rune, 0, len(runes))
	seen := make(map[rune]bool, len(runes))
	for _, r := range runes {
		if seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, r)
	}
	return out
}

// sattolo applies Sattolo's algorithm: the Fisher-Yates shuffle with the swap
// partner drawn from [0, i) instead of [0, i]. That draws a uniformly random
// single cycle, and a single cycle of length n ≥ 2 has no fixed point — every
// rune moves. It is therefore a derangement by construction, obtained in one
// pass with no rejection sampling.
//
// Two consequences are worth knowing. Every cycle is a derangement, but not
// every derangement is a cycle, so this samples a subset of them: the draw is
// uniform over single cycles, not over derangements, and the keyspace is
// (n-1)! rather than the !n a derangement drawn at large would allow. Neither
// matters for a puzzle, and it is why WithDerangement(false) exists for a
// caller who wants the whole space.
func sattolo(rng *rand.Rand, runes []rune) {
	for i := len(runes) - 1; i > 0; i-- {
		j := rng.IntN(i)
		runes[i], runes[j] = runes[j], runes[i]
	}
}

// isPermutation reports whether alphabet holds exactly the runes of charset,
// each once. It is what makes a cipher reversible.
func isPermutation(charset, alphabet []rune) bool {
	if len(charset) != len(alphabet) {
		return false
	}

	seen := make(map[rune]bool, len(alphabet))
	for _, r := range alphabet {
		if seen[r] {
			return false
		}
		seen[r] = true
	}
	for _, r := range charset {
		if !seen[r] {
			return false
		}
	}
	return true
}

// isDerangement reports whether no rune stands in for itself, comparing the
// two slices position by position.
func isDerangement(charset, alphabet []rune) bool {
	if len(charset) != len(alphabet) {
		return false
	}
	for i, plain := range charset {
		if alphabet[i] == plain {
			return false
		}
	}
	return true
}

// checkDistinct reports the first rune that appears twice, wrapped in
// ErrDuplicateRune, and nil when every rune is unique.
func checkDistinct(runes []rune) error {
	seen := make(map[rune]bool, len(runes))
	for _, r := range runes {
		if seen[r] {
			return fmt.Errorf("%w: %q appears twice", ErrDuplicateRune, r)
		}
		seen[r] = true
	}
	return nil
}

// decodeRunes reads one hex-encoded half of a key back into runes. The bytes
// must be valid UTF-8, because that is the only thing a charset can be built
// from: decoding invalid bytes would silently turn them into U+FFFD and hand
// back a cipher whose Key no longer matches the key it came from.
func decodeRunes(part string) ([]rune, error) {
	raw, err := hex.DecodeString(part)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("%w: %q is not valid UTF-8", ErrInvalidKey, part)
	}
	return []rune(string(raw)), nil
}

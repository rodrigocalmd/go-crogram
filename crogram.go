package crogram

import (
	"math/rand"
	"time"
)

// Cipher handles the encoding and decoding of text.
type Cipher struct {
	encodeMap map[rune]rune
	decodeMap map[rune]rune
}

// Encode transforms a string based on the Cipher's encoding map.
// Characters not in the encoding map are returned unchanged.
func (c *Cipher) Encode(text string) string {
	var result []rune
	for _, r := range text {
		if replacement, ok := c.encodeMap[r]; ok {
			result = append(result, replacement)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// Decode transforms a string based on the Cipher's decoding map.
// Characters not in the decoding map are returned unchanged.
func (c *Cipher) Decode(text string) string {
	var result []rune
	for _, r := range text {
		if original, ok := c.decodeMap[r]; ok {
			result = append(result, original)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// newCryptogram creates the substitution maps for the cipher.
func newCryptogram(r *rand.Rand) (map[rune]rune, map[rune]rune) {
	originalCharset := generateCharset()
	shuffledCharset := make([]rune, len(originalCharset))
	copy(shuffledCharset, originalCharset)
	r.Shuffle(len(shuffledCharset), func(i, j int) {
		shuffledCharset[i], shuffledCharset[j] = shuffledCharset[j], shuffledCharset[i]
	})

	encodeMap := make(map[rune]rune)
	decodeMap := make(map[rune]rune)

	for i, originalChar := range originalCharset {
		shuffledChar := shuffledCharset[i]
		encodeMap[originalChar] = shuffledChar
		decodeMap[shuffledChar] = originalChar
	}

	return encodeMap, decodeMap
}

// New creates a new Cipher.
// It can be called with an optional seed to generate a reproducible cipher.
// If no seed is provided, a random cipher is generated.
func New(seed ...int64) *Cipher {
	var r *rand.Rand
	if len(seed) > 0 {
		r = rand.New(rand.NewSource(seed[0]))
	} else {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	encodeMap, decodeMap := newCryptogram(r)

	return &Cipher{
		encodeMap: encodeMap,
		decodeMap: decodeMap,
	}
}

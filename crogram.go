package crogram

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"unicode/utf8"
)

// keyPrefix marks the portable, self-describing representation produced by
// Cipher.Key. The trailing version number lets a future format be detected
// rather than silently misread.
const keyPrefix = "crogram1:"

// Errors reported by ParseKey, NewFromPairs and NewFromMapping. They are
// wrapped with context, so test them with errors.Is.
var (
	// ErrInvalidKey reports a key that is not in the format Cipher.Key
	// produces.
	ErrInvalidKey = errors.New("crogram: invalid key")
	// ErrCharsetMismatch reports a key whose charset and alphabet hold a
	// different number of runes.
	ErrCharsetMismatch = errors.New("crogram: charset and alphabet differ in size")
	// ErrNotPermutation reports a key whose alphabet is not a rearrangement of
	// its charset, which would make the cipher impossible to reverse.
	ErrNotPermutation = errors.New("crogram: alphabet is not a permutation of the charset")
	// ErrDuplicateRune reports a character set or an alphabet that lists the
	// same rune twice.
	ErrDuplicateRune = errors.New("crogram: duplicate rune")
	// ErrUnpairedRune reports a mapping that pairs nothing up, such as an odd
	// number of runes passed to NewFromPairs.
	ErrUnpairedRune = errors.New("crogram: expected plaintext/replacement couples")
)

// Pair is one couple of a cipher: a plaintext rune and the rune that stands in
// for it.
type Pair struct {
	Plain  rune
	Cipher rune
}

// Cipher is a reversible substitution over a fixed character set.
//
// A Cipher holds two lookup tables built from its character set and a
// permutation of it, so Encode is a map lookup per rune and Decode is the same
// lookup against the reversed table. Runes outside the character set are
// copied through unchanged in both directions.
//
// A Cipher is immutable once built and therefore safe for concurrent use.
//
// See the package documentation for what this is not: a substitution cipher
// hides nothing from anyone who looks.
type Cipher struct {
	charset   []rune
	alphabet  []rune
	encodeMap map[rune]rune
	decodeMap map[rune]rune
	seed      int64
	seeded    bool
	deranged  bool
	// origin is the short account of where the cipher came from that String
	// reports. It never carries the mapping.
	origin string
}

// config carries the options New folds into a Cipher.
type config struct {
	charset  []rune
	seed     int64
	seeded   bool
	deranged bool
}

// Option configures New, NewWithSeed, NewFromPassphrase and NewPuzzle.
type Option func(*config)

// WithCharset replaces the default character set. Repeated runes are collapsed
// keeping the first occurrence, and an empty set leaves the default in place.
// The slice is copied, so the caller may keep mutating it afterwards.
func WithCharset(charset []rune) Option {
	return func(c *config) {
		if len(charset) > 0 {
			c.charset = slices.Clone(charset)
		}
	}
}

// WithSeed makes the cipher reproducible: the same seed always produces the
// same substitution. Prefer NewWithSeed(seed) when a seed is all you are
// passing, and see Cipher.Key for a representation that also survives across
// Go versions.
//
// The seed drives the PCG generator from math/rand/v2, so a seed carries 64
// bits and no more.
func WithSeed(seed int64) Option {
	return func(c *config) {
		c.seed = seed
		c.seeded = true
	}
}

// WithDerangement controls whether a rune may stand in for itself.
//
// It is on by default, which is what a cryptogram wants: if "e" encoded to "e"
// the puzzle would hand the solver a letter for free. Turn it off for the
// largest possible keyspace or to reproduce a mapping that allows fixed points.
//
// Derangement is enforced by drawing a single-cycle permutation, so it costs
// nothing and never loops.
func WithDerangement(enabled bool) Option {
	return func(c *config) {
		c.deranged = enabled
	}
}

// New builds a cipher. Called with no arguments it draws a fresh, random
// substitution; called with a seed it is reproducible:
//
//	c := crogram.New()     // random, never repeats
//	c := crogram.New(42)   // the same substitution every time
//
// This variadic form is kept from the original API. At most one seed is
// considered: New(1, 2) is New(1). NewCipher is the options-based constructor,
// and NewWithSeed, NewFromPassphrase and NewPuzzle are built on it.
func New(seed ...int64) *Cipher {
	if len(seed) > 0 {
		return NewWithSeed(seed[0])
	}
	return NewCipher()
}

// NewCipher builds a cipher from options, applied in order so that a later
// WithSeed wins over an earlier one. With no options it uses DefaultCharset, a
// seed drawn from math/rand/v2's global source, and derangement enabled:
//
//	c := crogram.NewCipher(
//		crogram.WithSeed(42),
//		crogram.WithCharset([]rune("abc")),
//		crogram.WithDerangement(false),
//	)
//
// NewCipher never fails and never panics: invalid configuration is either
// corrected (repeats collapsed, empty set replaced) or ignored.
func NewCipher(opts ...Option) *Cipher {
	cfg := config{charset: DefaultCharset, deranged: true}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	charset := sanitize(cfg.charset)
	alphabet := make([]rune, len(charset))
	copy(alphabet, charset)

	seed := seedFor(cfg)
	rng := rand.New(pcgFor(seed))
	if cfg.deranged {
		sattolo(rng, alphabet)
	} else {
		rng.Shuffle(len(alphabet), func(i, j int) {
			alphabet[i], alphabet[j] = alphabet[j], alphabet[i]
		})
	}

	// Report what the cipher is, not what was asked for: a one-rune character
	// set has nothing to move, so derangement is impossible there however the
	// option was set.
	origin := "random"
	if cfg.seeded {
		origin = fmt.Sprintf("seed %d", seed)
	}
	return fromAlphabet(charset, alphabet, seed, cfg.seeded, isDerangement(charset, alphabet), origin)
}

// NewWithSeed builds a reproducible cipher from seed. The same seed always
// yields the same substitution, in this program and in later sessions: the
// alphabet is drawn from math/rand/v2's PCG source, whose output for a given
// seed the standard library pins with a regression test. Treat that as a
// strong habit of the standard library rather than a contract, and use Key
// when the mapping has to outlive the process.
func NewWithSeed(seed int64, opts ...Option) *Cipher {
	return NewCipher(append([]Option{WithSeed(seed)}, opts...)...)
}

// NewFromPassphrase builds a reproducible cipher from a human-readable
// passphrase, which is friendlier than carrying a raw int64 around:
//
//	c := crogram.NewFromPassphrase("correct horse battery staple")
//
// The passphrase is hashed with SHA-256 and the first eight bytes become the
// seed, which is then used exactly as WithSeed would use it. Note the
// consequence: the cipher is exactly as guessable as the passphrase, and this
// is a substitution cipher either way — it protects nothing, it only repeats.
func NewFromPassphrase(passphrase string, opts ...Option) *Cipher {
	sum := sha256.Sum256([]byte(passphrase))
	seed := int64(binary.BigEndian.Uint64(sum[:8]))
	return NewCipher(append([]Option{WithSeed(seed)}, opts...)...)
}

// ParseKey rebuilds the cipher a key came from:
//
//	same, err := crogram.ParseKey(key)
//
// A key is self-describing — it carries the character set and the alphabet —
// and derives nothing from a seed, so it keeps working across Go versions and
// between programs and languages.
//
// ParseKey reports ErrInvalidKey, ErrCharsetMismatch or ErrNotPermutation for
// anything it cannot rebuild, which means a key from an untrusted source
// cannot produce an unusable cipher.
func ParseKey(key string) (*Cipher, error) {
	trimmed := strings.TrimSpace(key)
	if !strings.HasPrefix(trimmed, keyPrefix) {
		return nil, fmt.Errorf("%w: missing %q prefix", ErrInvalidKey, keyPrefix)
	}

	body := trimmed[len(keyPrefix):]
	charsetPart, alphabetPart, found := strings.Cut(body, ":")
	if !found {
		return nil, fmt.Errorf("%w: expected charset and alphabet separated by %q", ErrInvalidKey, ":")
	}

	charset, err := decodeRunes(charsetPart)
	if err != nil {
		return nil, err
	}
	alphabet, err := decodeRunes(alphabetPart)
	if err != nil {
		return nil, err
	}

	if len(charset) == 0 {
		return nil, fmt.Errorf("%w: empty charset", ErrInvalidKey)
	}
	if err := checkDistinct(charset); err != nil {
		return nil, err
	}
	if err := checkDistinct(alphabet); err != nil {
		return nil, err
	}
	if len(charset) != len(alphabet) {
		return nil, fmt.Errorf("%w: %d charset runes against %d alphabet runes",
			ErrCharsetMismatch, len(charset), len(alphabet))
	}
	if !isPermutation(charset, alphabet) {
		return nil, ErrNotPermutation
	}

	return fromAlphabet(charset, alphabet, 0, false, isDerangement(charset, alphabet), "key"), nil
}

// MustParseKey is ParseKey for keys known to be valid. It panics on a bad key
// and exists for package-level variables and tests.
func MustParseKey(key string) *Cipher {
	cipher, err := ParseKey(key)
	if err != nil {
		panic(err)
	}
	return cipher
}

// fromAlphabet is the single place a Cipher is assembled, so every constructor
// produces the same tables. origin is the account of where the cipher came from
// that String reports.
func fromAlphabet(charset, alphabet []rune, seed int64, seeded, deranged bool, origin string) *Cipher {
	c := &Cipher{
		charset:   slices.Clone(charset),
		alphabet:  slices.Clone(alphabet),
		encodeMap: make(map[rune]rune, len(charset)),
		decodeMap: make(map[rune]rune, len(charset)),
		seed:      seed,
		seeded:    seeded,
		deranged:  deranged,
		origin:    origin,
	}
	for i, plain := range c.charset {
		c.encodeMap[plain] = c.alphabet[i]
		c.decodeMap[c.alphabet[i]] = plain
	}
	return c
}

// seedFor resolves the seed to use, drawing one from math/rand/v2's global
// source when the caller did not supply one. That source is seeded randomly by
// the runtime, so two ciphers built in the same nanosecond still differ, which
// a clock reading cannot promise on every platform.
func seedFor(cfg config) int64 {
	if cfg.seeded {
		return cfg.seed
	}
	return rand.Int64()
}

// pcgFor turns the one int64 seed into the pair of words math/rand/v2's PCG
// source is built from. The second word is derived from the first with mix64,
// so a seed stays a single number and NewWithSeed(Seed()) still rebuilds the
// cipher it came from.
func pcgFor(seed int64) *rand.PCG {
	return rand.NewPCG(uint64(seed), mix64(uint64(seed)))
}

// mix64 is the splitmix64 finalizer. It spreads the bits of one word over
// another so that the two halves of a PCG state are not near copies of each
// other, which keeps seed 0 from being a special case.
func mix64(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}

// Encode replaces every rune of the character set in text with its
// replacement, and copies every other rune through untouched:
//
//	c := crogram.NewWithSeed(42)
//	c.Encode("Hello, world 123!")   // "mZccf, lf8cB KIs!" for seed 42
//
// The literal above is the value seed 42 draws today, from math/rand/v2's PCG
// source; it changes if that source ever changes, which the standard library
// works not to do. The promise Encode does keep for any cipher is the one
// below.
//
// Empty input gives an empty result. Invalid UTF-8 is copied through byte for
// byte, so Decode(Encode(text)) == text holds for any string at all. Encode
// allocates once and is safe for concurrent use.
func (c *Cipher) Encode(text string) string {
	return c.substitute(text, c.encodeMap)
}

// Decode reverses Encode, turning a ciphertext back into its plaintext. Runes
// outside the alphabet are copied through, exactly as Encode leaves runes
// outside the character set alone.
func (c *Cipher) Decode(text string) string {
	return c.substitute(text, c.decodeMap)
}

// DecodeWith decodes text while letting the caller override individual
// mappings — the move a solver makes when they have guessed letters:
//
//	known := map[rune]rune{'X': 'H', 'q': 'e'}
//	c.DecodeWith(encoded, known)
//
// Every rune present in known is decoded through it, everything else through
// the cipher, and runes held by neither are copied through. An empty known map
// is the same as calling Decode.
//
// Like Encode and Decode, DecodeWith copies invalid UTF-8 byte by byte instead
// of through the rune error, so arbitrary input survives it unchanged.
func (c *Cipher) DecodeWith(text string, known map[rune]rune) string {
	if len(known) == 0 {
		return c.Decode(text)
	}
	if text == "" {
		return ""
	}

	out := make([]byte, 0, len(text))
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && size == 1 {
			out = append(out, text[i])
			i++
			continue
		}
		if plain, ok := known[r]; ok {
			out = utf8.AppendRune(out, plain)
			i += size
			continue
		}
		if plain, ok := c.decodeMap[r]; ok {
			out = utf8.AppendRune(out, plain)
			i += size
			continue
		}
		out = append(out, text[i:i+size]...)
		i += size
	}
	return string(out)
}

// substitute walks text once, replacing whatever the table holds and copying
// whatever it does not.
//
// It walks bytes rather than ranging over runes on purpose. A range would turn
// every invalid UTF-8 byte into U+FFFD and write that back as three bytes, so
// Decode(Encode(text)) would stop matching text; copying the offending byte
// through unchanged keeps the round trip exact for any input, valid UTF-8 or
// not.
func (c *Cipher) substitute(text string, table map[rune]rune) string {
	if text == "" {
		return ""
	}

	out := make([]byte, 0, len(text))
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && size == 1 {
			out = append(out, text[i])
			i++
			continue
		}
		if replacement, ok := table[r]; ok {
			out = utf8.AppendRune(out, replacement)
			i += size
			continue
		}
		out = append(out, text[i:i+size]...)
		i += size
	}
	return string(out)
}

// Replacement returns the rune Encode substitutes for plain, and reports
// whether plain is part of the character set at all.
func (c *Cipher) Replacement(plain rune) (rune, bool) {
	replacement, ok := c.encodeMap[plain]
	return replacement, ok
}

// Original returns the rune Decode substitutes for cipher, and reports whether
// cipher is part of the alphabet at all.
func (c *Cipher) Original(cipher rune) (rune, bool) {
	plain, ok := c.decodeMap[cipher]
	return plain, ok
}

// Charset returns a copy of the character set, in its configured order.
func (c *Cipher) Charset() []rune {
	return slices.Clone(c.charset)
}

// Alphabet returns a copy of the alphabet — the runes that stand in for the
// character set, positionally matched to it.
func (c *Cipher) Alphabet() []rune {
	return slices.Clone(c.alphabet)
}

// Mapping returns a copy of the plaintext-to-replacement table, for debugging,
// printing or feeding NewFromMapping.
func (c *Cipher) Mapping() map[rune]rune {
	out := make(map[rune]rune, len(c.encodeMap))
	for plain, replacement := range c.encodeMap {
		out[plain] = replacement
	}
	return out
}

// Pairs returns every couple of the cipher, ordered by the character set.
func (c *Cipher) Pairs() []Pair {
	pairs := make([]Pair, 0, len(c.charset))
	for i, plain := range c.charset {
		pairs = append(pairs, Pair{Plain: plain, Cipher: c.alphabet[i]})
	}
	return pairs
}

// Hint reveals one couple of the cipher at random — the next thing a solver
// wants when they are stuck. It reports false only when the cipher holds no
// runes, and it draws from the package's own source when rng is nil.
func (c *Cipher) Hint(rng *rand.Rand) (Pair, bool) {
	n := len(c.charset)
	if n == 0 {
		return Pair{}, false
	}

	i := rand.IntN(n)
	if rng != nil {
		i = rng.IntN(n)
	}
	return Pair{Plain: c.charset[i], Cipher: c.alphabet[i]}, true
}

// Seed returns the seed the cipher was built from, and whether that seed was
// supplied explicitly.
//
// Only a cipher built by New, NewCipher, NewWithSeed or NewFromPassphrase
// carries a seed, and for those NewWithSeed(seed) rebuilds it even when the
// second result is false: an unseeded cipher still reports the seed it drew.
// A cipher rebuilt from a key, or built from explicit pairs or a mapping, has
// no seed at all — it reports (0, false), and NewWithSeed(0) is a different
// cipher. Carry those around with Key instead.
func (c *Cipher) Seed() (int64, bool) {
	return c.seed, c.seeded
}

// IsDeranged reports whether no rune of the character set stands in for itself,
// which is the property a cryptogram wants. It describes the mapping the cipher
// actually holds: a cipher built with derangement enabled is deranged except
// for a one-rune character set, where the property is impossible, and a cipher
// parsed from a key reports whatever that key contains.
func (c *Cipher) IsDeranged() bool {
	return c.deranged
}

// Len returns the number of runes in the character set.
func (c *Cipher) Len() int {
	return len(c.charset)
}

// Key returns a portable, self-describing representation of the cipher:
//
//	c := crogram.NewWithSeed(42)
//	key := c.Key()
//	same, err := crogram.ParseKey(key)
//
// Unlike a seed, a key carries the whole character set and the whole alphabet,
// so it rebuilds the cipher without relying on a random number generator, a Go
// version or even this language. Treat it as the answer sheet: anyone holding
// the key reads every message the cipher was used on.
func (c *Cipher) Key() string {
	var b strings.Builder
	b.WriteString(keyPrefix)
	b.WriteString(hex.EncodeToString([]byte(string(c.charset))))
	b.WriteByte(':')
	b.WriteString(hex.EncodeToString([]byte(string(c.alphabet))))
	return b.String()
}

// String describes the shape of the cipher without revealing the alphabet, so
// it is safe to log. It also says where the cipher came from: a random draw, a
// seed, a key, explicit pairs or a mapping.
func (c *Cipher) String() string {
	return fmt.Sprintf("crogram.Cipher{charset: %d runes, deranged: %t, %s}",
		len(c.charset), c.deranged, c.origin)
}

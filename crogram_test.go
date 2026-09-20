package crogram_test

import (
	"errors"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/rodrigocalmd/go-crogram"
)

// TestEncodeDecodeRoundTrip is the property the whole package rests on:
// decoding an encoded text returns the original text, whatever it contains.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	c := crogram.NewWithSeed(1)

	texts := []string{
		"",
		"Hello World 123!",
		"aB3",
		strings.Repeat("The quick brown fox. ", 40),
		"punctuation only: !!! ??? ,,, ;;;",
		"accentuated and multi-byte: ação, coração, 日本語",
		"\n\ttabs and\nnewlines\r\n",
	}
	for _, text := range texts {
		if got := c.Decode(c.Encode(text)); got != text {
			t.Errorf("Decode(Encode(%q)) = %q", text, got)
		}
	}
}

// TestEncodeIsDeterministicForASeed pins the reproducibility promise.
func TestEncodeIsDeterministicForASeed(t *testing.T) {
	const text = "Hello World 123!"

	first := crogram.NewWithSeed(42).Encode(text)
	for i := 0; i < 20; i++ {
		if got := crogram.NewWithSeed(42).Encode(text); got != first {
			t.Fatalf("seed 42 produced %q and then %q", first, got)
		}
	}

	if other := crogram.NewWithSeed(43).Encode(text); other == first {
		t.Errorf("seeds 42 and 43 agree on %q", first)
	}
}

// TestUnseededCiphersDiffer checks that New() really draws a fresh seed.
func TestUnseededCiphersDiffer(t *testing.T) {
	const text = "Hello World 123!"

	seen := make(map[string]bool, 8)
	for i := 0; i < 8; i++ {
		seen[crogram.New().Encode(text)] = true
	}
	if len(seen) < 2 {
		t.Errorf("eight random ciphers produced only %d distinct texts", len(seen))
	}
}

// TestPreservesRunesOutsideCharset documents the pass-through rule with a
// charset small enough to reason about by hand.
func TestPreservesRunesOutsideCharset(t *testing.T) {
	c := crogram.NewWithSeed(5, crogram.WithCharset([]rune("abc")))

	const text = "a b c! ?x"
	encoded := []rune(c.Encode(text))
	want := []rune(text)

	if len(encoded) != len(want) {
		t.Fatalf("Encode(%q) = %q, the rune count changed", text, string(encoded))
	}
	for i, r := range want {
		if r == 'a' || r == 'b' || r == 'c' {
			continue // part of the charset, expected to change
		}
		if encoded[i] != r {
			t.Errorf("Encode(%q) = %q: rune %d (%q) should have been preserved",
				text, string(encoded), i, r)
		}
	}
}

// TestEveryCharsetRuneIsReplaced checks that the default cipher leaves nothing
// where it was, which is the derangement promise.
func TestEveryCharsetRuneIsReplaced(t *testing.T) {
	c := crogram.NewWithSeed(9)

	for _, r := range c.Charset() {
		encoded := []rune(c.Encode(string(r)))
		if len(encoded) != 1 {
			t.Fatalf("Encode(%q) produced %d runes, want 1", r, len(encoded))
		}
		if encoded[0] == r {
			t.Errorf("%q stands in for itself", r)
		}
	}
}

// TestMappingIsABijection checks that the substitution is reversible: no two
// plaintext runes share a replacement.
func TestMappingIsABijection(t *testing.T) {
	c := crogram.NewWithSeed(4)

	used := make(map[rune]rune, c.Len())
	for plain, replacement := range c.Mapping() {
		if other, taken := used[replacement]; taken {
			t.Fatalf("%q and %q both encode to %q", other, plain, replacement)
		}
		used[replacement] = plain
	}
	if len(used) != c.Len() {
		t.Errorf("the alphabet holds %d distinct runes, want %d", len(used), c.Len())
	}
}

// TestDefaultCipherIsADerangement checks the Sattolo-based guarantee across a
// range of seeds, not just one lucky draw.
func TestDefaultCipherIsADerangement(t *testing.T) {
	for seed := int64(0); seed < 32; seed++ {
		c := crogram.NewWithSeed(seed)

		if !c.IsDeranged() {
			t.Fatalf("seed %d: IsDeranged() = false", seed)
		}
		for _, pair := range c.Pairs() {
			if pair.Plain == pair.Cipher {
				t.Fatalf("seed %d: %q stands in for itself", seed, pair.Plain)
			}
		}
	}
}

// TestCharsetIsCleanedUp covers the two corrections New makes rather than
// failing: repeats are collapsed and an empty set falls back to the default.
func TestCharsetIsCleanedUp(t *testing.T) {
	doubled := crogram.NewWithSeed(1, crogram.WithCharset([]rune("abcabc")))
	if got := doubled.Len(); got != 3 {
		t.Errorf("Len() = %d for a charset of three distinct runes, want 3", got)
	}
	if got := []rune(doubled.Encode("abcd")); len(got) != 4 || got[3] != 'd' {
		t.Errorf("Encode(abcd) = %q, want the trailing d passed through", string(got))
	}

	for _, charset := range [][]rune{nil, {}} {
		fallback := crogram.NewWithSeed(1, crogram.WithCharset(charset))
		if got, want := fallback.Len(), len(crogram.DefaultCharset); got != want {
			t.Errorf("Len() = %d for an empty charset, want the default %d", got, want)
		}
	}
}

// TestSingleRuneCharsetCannotBeDeranged checks that a cipher reports what it
// actually is: a one-rune character set has nothing to move, so derangement is
// impossible there however WithDerangement was set.
func TestSingleRuneCharsetCannotBeDeranged(t *testing.T) {
	c := crogram.NewWithSeed(1, crogram.WithCharset([]rune("a")))
	if c.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", c.Len())
	}
	if c.IsDeranged() {
		t.Error("a one-rune character set cannot be deranged, but IsDeranged() = true")
	}
	if got := c.Encode("a"); got != "a" {
		t.Errorf("Encode(a) = %q, want the only rune of the set", got)
	}

	// Two runes are enough for the promise to hold again.
	two := crogram.NewWithSeed(1, crogram.WithCharset([]rune("ab")))
	if !two.IsDeranged() {
		t.Error("IsDeranged() = false for a two-rune charset, where a swap is possible")
	}
}

// TestCipherReportsWhereItCameFrom checks that Seed and String describe the
// cipher truthfully: a seeded cipher replays from its seed, and a cipher that
// has no seed says so rather than offering one that rebuilds something else.
func TestCipherReportsWhereItCameFrom(t *testing.T) {
	seeded := crogram.NewWithSeed(42)
	if got := seeded.String(); !strings.Contains(got, "seed 42") {
		t.Errorf("String() = %q, want it to name the seed", got)
	}
	if seed, ok := seeded.Seed(); !ok || seed != 42 {
		t.Errorf("Seed() = (%d, %t), want (42, true)", seed, ok)
	}

	if got := crogram.New().String(); !strings.Contains(got, "random") {
		t.Errorf("String() = %q, want it to say the cipher was drawn at random", got)
	}

	parsed, err := crogram.ParseKey(seeded.Key())
	if err != nil {
		t.Fatalf("ParseKey: %v", err)
	}
	if got := parsed.String(); !strings.Contains(got, "key") {
		t.Errorf("String() = %q, want it to say the cipher came from a key", got)
	}
	if seed, ok := parsed.Seed(); ok || seed != 0 {
		t.Errorf("Seed() = (%d, %t) for a cipher rebuilt from a key, want (0, false)", seed, ok)
	}
	if crogram.NewWithSeed(0).Key() == parsed.Key() {
		t.Error("seed 0 rebuilt a cipher that actually came from a key")
	}

	pairs, err := crogram.NewFromPairs('a', 'q')
	if err != nil {
		t.Fatalf("NewFromPairs: %v", err)
	}
	if got := pairs.String(); !strings.Contains(got, "pairs") {
		t.Errorf("String() = %q, want it to say the cipher came from explicit pairs", got)
	}

	mapped, err := crogram.NewFromMapping(map[rune]rune{'a': 'q'})
	if err != nil {
		t.Fatalf("NewFromMapping: %v", err)
	}
	if got := mapped.String(); !strings.Contains(got, "mapping") {
		t.Errorf("String() = %q, want it to say the cipher came from a mapping", got)
	}
}

// TestEmptyDefaultCharsetFallsBackToTheBuiltin covers the case the docs invite:
// a program replaces DefaultCharset globally, then empties it. The cipher must
// still be usable, and above all its own key must parse back.
func TestEmptyDefaultCharsetFallsBackToTheBuiltin(t *testing.T) {
	saved := crogram.DefaultCharset
	t.Cleanup(func() { crogram.DefaultCharset = saved })

	for _, charset := range [][]rune{nil, {}} {
		crogram.DefaultCharset = charset

		c := crogram.NewWithSeed(1)
		if c.Len() == 0 {
			t.Fatalf("DefaultCharset = %v left the cipher with no runes", charset)
		}
		if _, err := crogram.ParseKey(c.Key()); err != nil {
			t.Errorf("the key of a cipher built from an empty default does not parse back: %v", err)
		}
	}
}

// TestNewKeepsTheVariadicSeedForm guards the original API: New() and New(seed)
// must keep working next to the options-based constructor.
func TestNewKeepsTheVariadicSeedForm(t *testing.T) {
	if got, want := crogram.New(12345).Key(), crogram.NewWithSeed(12345).Key(); got != want {
		t.Errorf("New(seed) and NewWithSeed(seed) disagree:\n%q\n%q", got, want)
	}
	if got, want := crogram.New(7).Encode("Hello"), crogram.NewWithSeed(7).Encode("Hello"); got != want {
		t.Errorf("New(7) encodes %q, want %q", got, want)
	}
	// Two ciphers drawn back to back must not agree: an unseeded New takes its
	// seed from a random source rather than reading the clock, so even two
	// calls in the same nanosecond differ. Written with variables because
	// staticcheck reads the text of the two calls as identical expressions.
	first, second := crogram.New(), crogram.New()
	if first.Key() == second.Key() {
		t.Error("New() produced the same cipher twice")
	}
}

// TestKeyRebuildsTheCipher is the portability promise: a key alone recreates
// an identical cipher, for seeded, passphrase and custom-charset ciphers alike.
func TestKeyRebuildsTheCipher(t *testing.T) {
	const text = "Hello World 123!"

	ciphers := []*crogram.Cipher{
		crogram.NewWithSeed(7),
		crogram.NewFromPassphrase("correct horse battery staple"),
		crogram.NewWithSeed(3, crogram.WithCharset([]rune("abcdefghij"))),
		crogram.NewWithSeed(3, crogram.WithDerangement(false)),
	}
	for _, c := range ciphers {
		rebuilt, err := crogram.ParseKey(c.Key())
		if err != nil {
			t.Fatalf("ParseKey(%q): %v", c.Key(), err)
		}
		if got, want := rebuilt.Encode(text), c.Encode(text); got != want {
			t.Errorf("the rebuilt cipher encodes %q, want %q", got, want)
		}
		if got, want := rebuilt.Len(), c.Len(); got != want {
			t.Errorf("the rebuilt cipher holds %d runes, want %d", got, want)
		}
		if got, want := rebuilt.Key(), c.Key(); got != want {
			t.Errorf("Key() round-tripped to %q, want %q", got, want)
		}
	}
}

// TestParseKeyAcceptsAnIdentityMapping checks that a key describing a mapping
// with fixed points is rebuilt as such rather than rejected.
func TestParseKeyAcceptsAnIdentityMapping(t *testing.T) {
	c, err := crogram.ParseKey("crogram1:6162:6162")
	if err != nil {
		t.Fatalf("ParseKey: %v", err)
	}
	if got := c.Encode("ab"); got != "ab" {
		t.Errorf("Encode(ab) = %q, want %q", got, "ab")
	}
	if c.IsDeranged() {
		t.Error("an identity mapping was reported as a derangement")
	}
}

// TestParseKeyRejectsBadKeys walks the ways a key can be wrong.
func TestParseKeyRejectsBadKeys(t *testing.T) {
	cases := []struct {
		name string
		key  string
		want error
	}{
		{"empty", "", crogram.ErrInvalidKey},
		{"unknown prefix", "crogram:6162:6162", crogram.ErrInvalidKey},
		{"missing separator", "crogram1:6162", crogram.ErrInvalidKey},
		{"empty charset", "crogram1::61", crogram.ErrInvalidKey},
		{"invalid hex", "crogram1:6w:6162", crogram.ErrInvalidKey},
		{"size mismatch", "crogram1:6162:616263", crogram.ErrCharsetMismatch},
		{"not a permutation", "crogram1:6162:6163", crogram.ErrNotPermutation},
		{"repeated charset rune", "crogram1:6161:6162", crogram.ErrDuplicateRune},
		{"repeated alphabet rune", "crogram1:6162:6161", crogram.ErrDuplicateRune},
		{"bytes that are not utf-8", "crogram1:ff:ff", crogram.ErrInvalidKey},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := crogram.ParseKey(tc.key)
			if err == nil {
				t.Fatalf("ParseKey(%q) = %v, want an error", tc.key, c)
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("ParseKey(%q) = %v, want %v", tc.key, err, tc.want)
			}
		})
	}
}

// TestExplicitMappings covers NewFromPairs and NewFromMapping, including the
// fact that a replacement need not belong to the character set.
func TestExplicitMappings(t *testing.T) {
	c, err := crogram.NewFromPairs('a', '1', 'b', '2', 'c', '3')
	if err != nil {
		t.Fatalf("NewFromPairs: %v", err)
	}
	if got := c.Encode("abc"); got != "123" {
		t.Errorf("Encode(abc) = %q, want %q", got, "123")
	}
	if got := c.Decode("123"); got != "abc" {
		t.Errorf("Decode(123) = %q, want %q", got, "abc")
	}
	if got := c.Encode("abcd"); got != "123d" {
		t.Errorf("Encode(abcd) = %q, want the trailing d passed through", got)
	}

	odd := []struct {
		name string
		err  error
		run  func() error
	}{
		{"unpaired", crogram.ErrUnpairedRune, func() error {
			_, err := crogram.NewFromPairs('a')
			return err
		}},
		{"repeated plaintext", crogram.ErrDuplicateRune, func() error {
			_, err := crogram.NewFromPairs('a', '1', 'a', '2')
			return err
		}},
		{"repeated replacement", crogram.ErrDuplicateRune, func() error {
			_, err := crogram.NewFromPairs('a', '1', 'b', '1')
			return err
		}},
		{"empty mapping", crogram.ErrUnpairedRune, func() error {
			_, err := crogram.NewFromMapping(nil)
			return err
		}},
	}
	for _, tc := range odd {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); !errors.Is(err, tc.err) {
				t.Errorf("got %v, want %v", err, tc.err)
			}
		})
	}

	rebuilt, err := crogram.NewFromMapping(c.Mapping())
	if err != nil {
		t.Fatalf("NewFromMapping: %v", err)
	}
	if got, want := rebuilt.Encode("abc"), c.Encode("abc"); got != want {
		t.Errorf("NewFromMapping(Mapping()) encodes %q, want %q", got, want)
	}
}

// TestDecodeWith covers the solver's override path.
func TestDecodeWith(t *testing.T) {
	c := crogram.NewWithSeed(2, crogram.WithCharset([]rune("abc")))

	encoded := c.Encode("abc")
	if len([]rune(encoded)) != 3 {
		t.Fatalf("Encode(abc) = %q, want three runes", encoded)
	}

	known := map[rune]rune{[]rune(encoded)[0]: 'a'}
	if got := c.DecodeWith(encoded, known); !strings.HasPrefix(got, "a") {
		t.Errorf("DecodeWith with a guessed first rune gave %q, want a leading a", got)
	}
	if got, want := c.DecodeWith(encoded, nil), c.Decode(encoded); got != want {
		t.Errorf("DecodeWith with no guesses gave %q, want %q", got, want)
	}
	if got := c.DecodeWith("! ?", nil); got != "! ?" {
		t.Errorf("DecodeWith copied %q through as %q, want it unchanged", "! ?", got)
	}
}

// TestFrequencyAndSuggestion covers the analysis helpers.
func TestFrequencyAndSuggestion(t *testing.T) {
	f := crogram.Analyse("aab")

	if f.Total != 3 {
		t.Errorf("Total = %d, want 3", f.Total)
	}
	if got := f.Counts['a']; got != 2 {
		t.Errorf("Counts[a] = %d, want 2", got)
	}
	if got, want := f.Rate('a'), 2.0/3.0; got != want {
		t.Errorf("Rate(a) = %v, want %v", got, want)
	}
	if got := f.Rate('z'); got != 0 {
		t.Errorf("Rate(z) = %v, want 0", got)
	}
	if sorted := f.Sorted(); len(sorted) != 2 || sorted[0].Rune != 'a' || sorted[0].Count != 2 {
		t.Errorf("Sorted() = %+v, want a first and then b", sorted)
	}
	if got := f.Top(1); len(got) != 1 || got[0] != 'a' {
		t.Errorf("Top(1) = %v, want [a]", got)
	}
	if got := f.Top(-1); len(got) != 0 {
		t.Errorf("Top(-1) = %v, want nothing", got)
	}
	if got := crogram.Analyse("").Rate('a'); got != 0 {
		t.Errorf("Rate on an empty profile = %v, want 0", got)
	}

	c := crogram.NewWithSeed(13)
	plain := strings.Repeat("the quick brown fox jumps over the lazy dog ", 4)
	guess := c.SuggestedGuess(c.Encode(plain))

	if len(guess) == 0 {
		t.Fatal("SuggestedGuess produced no guesses for a long text")
	}
	if len(guess) > len(crogram.EnglishFrequency) {
		t.Errorf("SuggestedGuess made %d guesses, more than the %d reference letters",
			len(guess), len(crogram.EnglishFrequency))
	}
	alphabet := make(map[rune]bool, c.Len())
	for _, r := range c.Alphabet() {
		alphabet[r] = true
	}
	for cipherRune := range guess {
		if !alphabet[cipherRune] {
			t.Errorf("SuggestedGuess keyed on %q, which is not in the alphabet", cipherRune)
		}
	}
}

// TestHintRevealsRealCouples checks that a hint is always a true statement
// about the cipher.
func TestHintRevealsRealCouples(t *testing.T) {
	c := crogram.NewWithSeed(11)
	rng := rand.New(rand.NewPCG(1, 2))

	seen := make(map[crogram.Pair]bool, c.Len())
	for i := 0; i < 200; i++ {
		pair, ok := c.Hint(rng)
		if !ok {
			t.Fatal("Hint reported no couple for a 62-rune cipher")
		}
		if got := c.Encode(string(pair.Plain)); got != string(pair.Cipher) {
			t.Fatalf("Hint gave %q -> %q, but Encode says %q", pair.Plain, pair.Cipher, got)
		}
		seen[pair] = true
	}
	if len(seen) < 2 {
		t.Errorf("200 hints revealed %d distinct couples, want more", len(seen))
	}
	if _, ok := c.Hint(nil); !ok {
		t.Error("Hint(nil) should fall back to a source of its own")
	}
}

// TestSeedAccessors checks the replay promise: a cipher that was not seeded
// still reports the seed that rebuilds it.
func TestSeedAccessors(t *testing.T) {
	seeded := crogram.NewWithSeed(42)
	if seed, ok := seeded.Seed(); !ok || seed != 42 {
		t.Errorf("Seed() = (%d, %t), want (42, true)", seed, ok)
	}

	random := crogram.New()
	seed, ok := random.Seed()
	if ok {
		t.Error("a random cipher reported an explicit seed")
	}
	if got := crogram.NewWithSeed(seed).Encode("abc"); got != random.Encode("abc") {
		t.Errorf("replaying the reported seed gave %q, want %q", got, random.Encode("abc"))
	}
}

// TestPassphraseCiphersAreReproducible covers the SHA-256 derivation.
func TestPassphraseCiphersAreReproducible(t *testing.T) {
	a := crogram.NewFromPassphrase("correct horse battery staple")
	b := crogram.NewFromPassphrase("correct horse battery staple")
	if a.Key() != b.Key() {
		t.Error("the same passphrase produced two different ciphers")
	}

	c := crogram.NewFromPassphrase("correct horse battery stapl")
	if c.Key() == a.Key() {
		t.Error("a different passphrase produced the same cipher")
	}
	if _, ok := a.Seed(); !ok {
		t.Error("a passphrase cipher should report an explicit seed")
	}
}

// TestPuzzle covers the puzzle front end.
func TestPuzzle(t *testing.T) {
	puzzle := crogram.NewPuzzle("The quick brown fox jumps over the lazy dog", crogram.WithSeed(7))

	if puzzle.Ciphertext == puzzle.Plaintext {
		t.Error("NewPuzzle did not encode the plaintext")
	}
	if !puzzle.Solved(puzzle.Plaintext) {
		t.Error("Solved rejected the plaintext")
	}
	if puzzle.Solved(puzzle.Ciphertext) {
		t.Error("Solved accepted the ciphertext")
	}
	if got, want := puzzle.String(), puzzle.Ciphertext; got != want {
		t.Errorf("String() = %q, want the ciphertext %q", got, want)
	}
	if got, want := puzzle.Key(), puzzle.Cipher.Key(); got != want {
		t.Errorf("Key() = %q, want %q", got, want)
	}

	rebuilt, err := crogram.ParseKey(puzzle.Key())
	if err != nil {
		t.Fatalf("ParseKey(puzzle.Key()): %v", err)
	}
	if got := rebuilt.Decode(puzzle.Ciphertext); got != puzzle.Plaintext {
		t.Errorf("the puzzle key decodes to %q, want %q", got, puzzle.Plaintext)
	}

	rng := rand.New(rand.NewPCG(2, 3))
	if _, ok := puzzle.Hint(rng); !ok {
		t.Error("Hint reported no couple")
	}
}

// TestStringKeepsTheAlphabetToItself checks that logging a cipher does not leak
// the mapping.
func TestStringKeepsTheAlphabetToItself(t *testing.T) {
	c := crogram.NewWithSeed(9)
	got := c.String()

	if strings.Contains(got, string(c.Alphabet())) {
		t.Errorf("String() leaked the alphabet: %q", got)
	}
	if strings.Contains(got, c.Key()) {
		t.Errorf("String() leaked the key: %q", got)
	}
	if !strings.Contains(got, "62 runes") {
		t.Errorf("String() = %q, want it to mention the 62-rune charset", got)
	}
}

// TestInvalidUTF8SurvivesByteForByte covers the case the fuzzer found: a range
// over a string turns every invalid byte into U+FFFD and writes back three
// bytes, so the round trip has to walk bytes rather than runes.
func TestInvalidUTF8SurvivesByteForByte(t *testing.T) {
	c := crogram.NewWithSeed(1)

	for _, text := range []string{"\xb7", "a\xb7b", "\xff\xfe", "ok\xc3"} {
		encoded := c.Encode(text)
		if len(encoded) != len(text) {
			t.Errorf("Encode(%q) = %q: %d bytes became %d", text, encoded, len(text), len(encoded))
		}
		if got := c.Decode(encoded); got != text {
			t.Errorf("Decode(Encode(%q)) = %q", text, got)
		}
	}
}

// TestDocumentedSeedOutputIsStillTrue is the suite's one golden value, and it
// exists only to keep the prose honest: the README, doc.go and Cipher.Encode
// all quote what seed 42 encodes to, and a change in math/rand/v2's stream
// would leave those quotes quietly wrong. When this fails, update the two
// literals rather than the test.
func TestDocumentedSeedOutputIsStillTrue(t *testing.T) {
	c := crogram.NewWithSeed(42)

	if got, want := c.Encode("Hello World 123!"), "mZccf Df8cB KIs!"; got != want {
		t.Errorf("README quotes seed 42 encoding %q, but it encodes %q", want, got)
	}
	if got, want := c.Encode("Hello, world 123!"), "mZccf, lf8cB KIs!"; got != want {
		t.Errorf("Cipher.Encode quotes seed 42 encoding %q, but it encodes %q", want, got)
	}
	key := c.Key()
	if !strings.HasPrefix(key, "crogram1:61626364") {
		t.Errorf("README and doc.go quote a key starting %q, but it starts %q",
			"crogram1:61626364", key[:20])
	}
	if _, alphabet, found := strings.Cut(strings.TrimPrefix(key, "crogram1:"), ":"); !found || !strings.HasPrefix(alphabet, "69323742") {
		t.Errorf("README and doc.go quote an alphabet starting %q, but it starts %q",
			"69323742", alphabet[:8])
	}
}

// TestNonLatinCharsets checks that the cipher is not tied to ASCII: any rune
// set works as a character set, whether the script has case, writes right to
// left, or needs three bytes per letter.
func TestNonLatinCharsets(t *testing.T) {
	cases := []struct {
		name    string
		charset string
		text    string
	}{
		{"greek", "αβγδεζηθικλμνξοπρστυφχψωΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΤΥΦΧΨΩ", "καλημέρα"},
		{"cyrillic", "абвгдежзийклмнопрстуфхцчшщыьэюя", "привет"},
		{"arabic", "ابتثجحخدذرزسشصضطظعغفقكلمنهوي", "مرحبا"},
		{"hebrew", "שלוםעולםבוקרטי", "שלום"},
		{"japanese", "あいうえおかきくけこさしすせそたちつてと", "こんにちは"},
		{"emoji", "🍎🍌🍇🍓🍒🍑🥝🍍", "🍎🍌🍇🍓"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := crogram.NewWithSeed(3, crogram.WithCharset([]rune(tc.charset)))

			if got := c.Decode(c.Encode(tc.text)); got != tc.text {
				t.Errorf("Decode(Encode(%q)) = %q", tc.text, got)
			}
			for _, r := range tc.charset {
				if got := c.Encode(string(r)); got == string(r) {
					t.Errorf("%q stands in for itself", r)
				}
			}

			rebuilt, err := crogram.ParseKey(c.Key())
			if err != nil {
				t.Fatalf("ParseKey: %v", err)
			}
			if got, want := rebuilt.Encode(tc.text), c.Encode(tc.text); got != want {
				t.Errorf("the rebuilt cipher encodes %q, want %q", got, want)
			}
			if got, want := rebuilt.Charset(), c.Charset(); !slices.Equal(got, want) {
				t.Errorf("Charset() round-tripped to %q, want %q", string(got), string(want))
			}
		})
	}
}

// TestDefaultCharsetLeavesAccentsAlone documents the consequence of the default
// character set being ASCII: in a language written with accents, those letters
// are outside the set and pass through in clear, which hands a solver free
// clues. A charset that covers the language is what fixes it.
func TestDefaultCharsetLeavesAccentsAlone(t *testing.T) {
	const text = "ação"

	c := crogram.NewWithSeed(1)
	if _, ok := c.Replacement('ç'); ok {
		t.Fatal("the default charset should not contain ç")
	}

	// The ç and ã are not in the set, so they survive in place while the a and
	// the o — which are — do not. The rune count is unchanged either way,
	// because one rune is replaced by exactly one rune.
	encoded := c.Encode(text)
	if !strings.Contains(encoded, "çã") {
		t.Errorf("Encode(%q) = %q, want ç and ã passed through in clear", text, encoded)
	}
	if got, want := len([]rune(encoded)), len([]rune(text)); got != want {
		t.Errorf("Encode(%q) = %q: %d runes became %d", text, encoded, want, got)
	}

	accented := crogram.NewWithSeed(1, crogram.WithCharset([]rune("abcdefghijklmnopqrstuvwxyzçã")))
	for _, r := range text {
		if _, ok := accented.Replacement(r); !ok {
			t.Errorf("%q is not in the charset, so it would pass through", r)
		}
	}
	encoded = accented.Encode(text)
	if encoded == text {
		t.Errorf("Encode(%q) left the text unchanged", text)
	}
	if got := accented.Decode(encoded); got != text {
		t.Errorf("Decode(Encode(%q)) = %q", text, got)
	}
}

// TestCombiningMarksAreTheirOwnRunes pins the Unicode subtlety the docs warn
// about: a precomposed letter and the same letter spelled as base plus
// combining mark are different texts here, and encode differently, while both
// still round-trip. Normalising input is the caller's business, because doing
// it here would pull in a dependency this module does not have.
func TestCombiningMarksAreTheirOwnRunes(t *testing.T) {
	c := crogram.NewWithSeed(4, crogram.WithCharset([]rune("abcdefghijklmnopqrstuvwxyzé")))

	precomposed := "café"      // é as one rune, U+00E9
	decomposed := "cafe\u0301" // e followed by the combining acute

	if got := c.Encode(precomposed); got == c.Encode(decomposed) {
		t.Errorf("both spellings encoded to %q, want them to differ", got)
	}
	for _, text := range []string{precomposed, decomposed} {
		if got := c.Decode(c.Encode(text)); got != text {
			t.Errorf("Decode(Encode(%q)) = %q", text, got)
		}
	}
}

// TestConcurrentUse is meant to be run with -race: a Cipher is shared, not
// copied, and the whole point of its immutability is that this is safe.
func TestConcurrentUse(t *testing.T) {
	const text = "Hello World 123!"

	c := crogram.NewWithSeed(3)
	encoded := c.Encode(text)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 250; j++ {
				if got := c.Encode(text); got != encoded {
					t.Errorf("Encode in a goroutine = %q, want %q", got, encoded)
					return
				}
				if got := c.Decode(encoded); got != text {
					t.Errorf("Decode in a goroutine = %q, want %q", got, text)
					return
				}
				_ = c.Key()
				_ = c.Mapping()
			}
		}()
	}
	wg.Wait()
}

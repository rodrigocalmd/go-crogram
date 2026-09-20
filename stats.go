package crogram

import (
	"cmp"
	"slices"
)

// RuneCount pairs a rune with how many times it occurs.
type RuneCount struct {
	Rune  rune
	Count int
}

// Frequency is a letter-frequency profile of a text: how often each rune
// occurs, and how many runes were counted in total.
//
// Frequency analysis is how a substitution cipher is broken, which makes this
// the first tool a solver reaches for.
type Frequency struct {
	Total  int
	Counts map[rune]int
}

// Analyse counts every rune of text, including runes outside any cipher's
// character set, so the profile can be reused for the plaintext and the
// ciphertext of the same puzzle.
func Analyse(text string) Frequency {
	f := Frequency{Counts: make(map[rune]int)}
	for _, r := range text {
		f.Counts[r]++
		f.Total++
	}
	return f
}

// Rate returns the share of the analysed text made up by r, between 0 and 1.
// An empty profile rates everything at 0.
func (f Frequency) Rate(r rune) float64 {
	if f.Total == 0 {
		return 0
	}
	return float64(f.Counts[r]) / float64(f.Total)
}

// compareRuneCounts orders runes by how often they occur, most frequent first,
// and breaks ties by rune value, so a ranking is deterministic and safe to
// compare in tests.
func compareRuneCounts(a, b RuneCount) int {
	return cmp.Or(cmp.Compare(b.Count, a.Count), cmp.Compare(a.Rune, b.Rune))
}

// Sorted lists every rune that occurred, most frequent first. Runes with equal
// counts are ordered by rune value, so the result is deterministic and safe to
// compare in tests.
func (f Frequency) Sorted() []RuneCount {
	out := make([]RuneCount, 0, len(f.Counts))
	for r, count := range f.Counts {
		out = append(out, RuneCount{Rune: r, Count: count})
	}
	slices.SortFunc(out, compareRuneCounts)
	return out
}

// Top returns the n most frequent runes, or as many as occurred when n is
// larger. A negative n returns nothing.
func (f Frequency) Top(n int) []rune {
	sorted := f.Sorted()
	if n > len(sorted) {
		n = len(sorted)
	}
	if n < 0 {
		n = 0
	}

	out := make([]rune, 0, n)
	for _, rc := range sorted[:n] {
		out = append(out, rc.Rune)
	}
	return out
}

// EnglishFrequency is the classic descending order of English letters. It is
// the default stand-in when a substitution has to be guessed from frequencies
// alone, and a caller is free to substitute another language's order.
var EnglishFrequency = []rune("etaoinshrdlucmfwypvbgkjqxz")

// SuggestedGuess ranks the runes of the cipher's alphabet by how often they
// occur in ciphertext and pairs them, most frequent first, with
// EnglishFrequency:
//
//	guess := cipher.SuggestedGuess(encoded)
//	fmt.Println(cipher.DecodeWith(encoded, guess))
//
// The result is shaped for DecodeWith: its keys are ciphertext runes and its
// values are the plaintext runes guessed for them. It is a starting point, not
// an answer — letter order only holds for long, ordinary prose, and nothing
// here checks that the guess is self-consistent.
func (c *Cipher) SuggestedGuess(ciphertext string) map[rune]rune {
	freq := Analyse(ciphertext)

	ranked := make([]RuneCount, 0, len(c.alphabet))
	for _, r := range c.alphabet {
		if count, ok := freq.Counts[r]; ok {
			ranked = append(ranked, RuneCount{Rune: r, Count: count})
		}
	}
	slices.SortFunc(ranked, compareRuneCounts)

	guess := make(map[rune]rune, len(ranked))
	for i, rc := range ranked {
		if i >= len(EnglishFrequency) {
			break
		}
		guess[rc.Rune] = EnglishFrequency[i]
	}
	return guess
}

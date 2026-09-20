package crogram

import "math/rand/v2"

// Puzzle is a cryptogram challenge: the plaintext, the coded text a solver is
// shown, and the cipher that links the two.
type Puzzle struct {
	// Plaintext is what the puzzle hides.
	Plaintext string
	// Ciphertext is the coded text a solver is given.
	Ciphertext string
	// Cipher is the substitution that turns one into the other.
	Cipher *Cipher
}

// NewPuzzle encodes plaintext with a fresh cipher built from opts:
//
//	puzzle := crogram.NewPuzzle("The quick brown fox",
//		crogram.WithSeed(7))
//	fmt.Println(puzzle)        // the solver sees only the coded text
//	fmt.Println(puzzle.Key())  // the answer sheet
//
// Derangement is on by default, so no letter survives encoding as itself and
// the puzzle gives nothing away for free.
func NewPuzzle(plaintext string, opts ...Option) *Puzzle {
	return NewPuzzleWith(NewCipher(opts...), plaintext)
}

// NewPuzzleWith encodes plaintext with a cipher the caller already has, which
// is what a command line tool does after reading a key or a seed from its
// flags.
func NewPuzzleWith(cipher *Cipher, plaintext string) *Puzzle {
	return &Puzzle{
		Plaintext:  plaintext,
		Ciphertext: cipher.Encode(plaintext),
		Cipher:     cipher,
	}
}

// Key returns the key of the puzzle's cipher: the solution, not the puzzle.
// Hand it to ParseKey to rebuild the cipher and read the answer.
func (p *Puzzle) Key() string {
	return p.Cipher.Key()
}

// Hint reveals one substitution of the puzzle, which is what a stuck solver
// wants next. Pass a source of your own, such as rand.New(rand.NewPCG(1, 2)),
// or nil to draw from the package's own. See Cipher.Hint.
func (p *Puzzle) Hint(rng *rand.Rand) (Pair, bool) {
	return p.Cipher.Hint(rng)
}

// Solved reports whether guess is the puzzle's plaintext.
func (p *Puzzle) Solved(guess string) bool {
	return guess == p.Plaintext
}

// String returns the coded text a solver is shown. It never prints the
// plaintext, so a Puzzle can be logged or sent without giving the answer away.
func (p *Puzzle) String() string {
	return p.Ciphertext
}

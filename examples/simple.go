// Command simple demonstrates the crogram package: a random cipher, a
// reproducible one, a portable key, and a puzzle with its answer.
//
// Run it from the repository root:
//
//	go run ./examples
package main

import (
	"fmt"
	"math/rand/v2"

	"github.com/rodrigocalmd/go-crogram"
)

func main() {
	const original = "Hello World 123!"

	// A random cipher draws a fresh substitution every time it is built.
	randomCipher := crogram.New()
	encoded := randomCipher.Encode(original)
	fmt.Println("Encoded: ", encoded)
	fmt.Println("Decoded: ", randomCipher.Decode(encoded))

	// A seeded cipher repeats in this program and in any later session.
	seeded := crogram.NewWithSeed(42)
	fmt.Println("Seeded:  ", seeded.Encode(original))
	fmt.Println("Reproducible:", crogram.NewWithSeed(42).Key() == seeded.Key())

	// A key carries the whole cipher, so it survives across Go versions and
	// between programs. Treat it as the answer sheet.
	copied, err := crogram.ParseKey(seeded.Key())
	if err != nil {
		panic(err)
	}
	fmt.Println("From key:", copied.Encode(original))
	fmt.Println("Key:     ", seeded.Key())

	// A puzzle for a solver: the coded text comes out, the answer stays here.
	puzzle := crogram.NewPuzzle("The quick brown fox jumps over the lazy dog",
		crogram.WithSeed(7))
	fmt.Println("Puzzle:  ", puzzle)
	fmt.Println("Solved:  ", puzzle.Solved(puzzle.String()))

	// Hints are for when they get stuck. Passing nil draws from the package's
	// own source; a seeded one keeps a run of the example reproducible.
	rng := rand.New(rand.NewPCG(1, 2))
	if pair, ok := puzzle.Hint(rng); ok {
		fmt.Printf("Hint:     %q is really %q\n", pair.Cipher, pair.Plain)
	}

	// Frequency analysis is how such a puzzle is broken in the first place.
	profile := crogram.Analyse(puzzle.Ciphertext)
	fmt.Println("Commonest:", profile.Top(3))
	fmt.Println("First guess:")
	fmt.Println(puzzle.Cipher.DecodeWith(puzzle.Ciphertext,
		puzzle.Cipher.SuggestedGuess(puzzle.Ciphertext)))

	// Describe a cipher without giving it away.
	fmt.Println("Cipher:  ", seeded)
}

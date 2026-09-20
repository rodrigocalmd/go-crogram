package crogram_test

import (
	"fmt"

	"github.com/rodrigocalmd/go-crogram"
)

// Every example below prints something that is true of any substitution, not
// of one particular draw, so the output comments hold whatever the cipher is.

func ExampleNew() {
	c := crogram.NewWithSeed(42)

	encoded := c.Encode("Hello World 123!")
	fmt.Println(c.Decode(encoded))
	// Output: Hello World 123!
}

func ExampleNewWithSeed() {
	first := crogram.NewWithSeed(2026)
	second := crogram.NewWithSeed(2026)

	fmt.Println(first.Encode("meet me at midnight") == second.Encode("meet me at midnight"))
	// Output: true
}

func ExampleNewFromPassphrase() {
	c := crogram.NewFromPassphrase("correct horse battery staple")

	encoded := c.Encode("shibboleth")
	fmt.Println(encoded != "shibboleth")
	fmt.Println(c.Decode(encoded))
	// Output:
	// true
	// shibboleth
}

func ExampleNew_derangement() {
	c := crogram.NewWithSeed(42)

	fixed := 0
	for _, pair := range c.Pairs() {
		if pair.Plain == pair.Cipher {
			fixed++
		}
	}
	fmt.Println(fixed)
	// Output: 0
}

func ExampleCipher_Key() {
	c := crogram.NewWithSeed(42)

	same, err := crogram.ParseKey(c.Key())
	fmt.Println(err)
	fmt.Println(same.Encode("Hello World 123!") == c.Encode("Hello World 123!"))
	// Output:
	// <nil>
	// true
}

func ExampleCipher_DecodeWith() {
	c := crogram.NewWithSeed(7)
	encoded := c.Encode("hello")

	// A solver who has worked out that the first rune stands for 'h' can apply
	// that guess without decoding the rest by hand.
	guess := map[rune]rune{[]rune(encoded)[0]: 'h'}
	fmt.Println(c.DecodeWith(encoded, guess)[:1])
	// Output: h
}

func ExampleNewFromPairs() {
	c, err := crogram.NewFromPairs('a', 'q', 'b', 'x', 'c', 'm')
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(c.Encode("abc"))
	fmt.Println(c.Decode("qxm"))
	// Output:
	// qxm
	// abc
}

func ExampleAnalyse() {
	f := crogram.Analyse("aab")

	fmt.Println(f.Counts['a'], f.Total)
	// Output: 2 3
}

func ExampleNewPuzzle() {
	puzzle := crogram.NewPuzzle("The quick brown fox", crogram.WithSeed(11))

	fmt.Println(puzzle.Solved(puzzle.String()))
	fmt.Println(puzzle.Solved(puzzle.Plaintext))
	// Output:
	// false
	// true
}

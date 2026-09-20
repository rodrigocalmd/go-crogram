# go-crogram

[![Go Reference](https://pkg.go.dev/badge/github.com/rodrigocalmd/go-crogram.svg)](https://pkg.go.dev/github.com/rodrigocalmd/go-crogram)
[![CI](https://github.com/rodrigocalmd/go-crogram/actions/workflows/ci.yml/badge.svg)](https://github.com/rodrigocalmd/go-crogram/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A small, dependency-free Go package for building **substitution ciphers** — the
kind of letter-for-letter puzzle a cryptogram is — and for encoding, decoding
and solving text with them.

```go
c := crogram.NewWithSeed(42)

encoded := c.Encode("Hello World 123!")   // "mZccf Df8cB KIs!" for seed 42
c.Decode(encoded)                         // "Hello World 123!"
```

The encoded literal is what seed 42 draws today, from `math/rand/v2`'s PCG
source, so it is stable across Go releases but not guaranteed to be — which is
why `example_test.go` prints values that hold for any cipher instead.

## ⚠️ This is not cryptography

`crogram` is for puzzles, games, teaching and light obfuscation. A substitution
cipher leaks letter frequency, word lengths and repeated words, and a person
with a pencil — or [the frequency tools in this same package](#puzzles-and-solving) —
breaks it quickly. **Never use it to protect secrets, credentials or personal
data.** Reach for `crypto/aes` or `crypto/chacha20poly1305` and a real
key-derivation function for that.

## Features

- **Substitution over any character set.** The default covers `a-z`, `A-Z` and
  `0-9`; `WithCharset` replaces it, and a charset need not be ASCII or even
  Latin.
- **Everything else is preserved.** Spaces, punctuation, newlines and runes
  outside the set pass through untouched, so word shapes and layout survive.
- **No rune stands in for itself.** Derangement is on by default — a cryptogram
  that maps `e` to `e` would hand the solver a letter for free.
- **Reproducible.** `WithSeed` and `NewFromPassphrase` rebuild the same
  substitution in any later session.
- **Portable.** `Key` / `ParseKey` carry the whole cipher as a self-describing
  string that survives across Go versions, programs and languages.
- **Built for solvers.** Frequency analysis, partial decoding with guessed
  letters, hints, and a first-guess suggestion.
- **Puzzle front end.** `NewPuzzle` encodes a text and keeps the answer.
- **Safe to share.** A `Cipher` is immutable, so it is safe for concurrent use;
  `String` describes a cipher without revealing it.
- **No dependencies.** Standard library only.

## Install

```bash
go get github.com/rodrigocalmd/go-crogram
```

Requires Go 1.24 or newer, which is what this module declares. CI builds and
tests both that declared floor and the newest Go (1.27.1 at the time of
writing), so the version above is tested rather than guessed.

## Quick start

```go
package main

import (
	"fmt"

	"github.com/rodrigocalmd/go-crogram"
)

func main() {
	// A random cipher draws a fresh substitution every time.
	randomCipher := crogram.New()
	encoded := randomCipher.Encode("Hello World 123!")
	fmt.Println("Encoded:", encoded)
	fmt.Println("Decoded:", randomCipher.Decode(encoded))

	// A seeded cipher repeats, in this program and in later sessions.
	seeded := crogram.NewWithSeed(42)
	fmt.Println("Seeded:  ", seeded.Encode("Hello World 123!"))
}
```

## Reproducibility: seeds and keys

A **seed** is a number. `NewWithSeed(42)` always produces the same substitution,
which makes it good for tests and for recreating a ciphertext you did not store.
The alphabet is drawn from `math/rand/v2`'s PCG source, and the standard library
pins a given seed's output with a regression test; that is a strong habit of the
standard library rather than a documented contract, so keep a **key**, not a
seed, for anything that has to outlive the process.

```go
seed, explicit := cipher.Seed()   // the seed it was built from
```

A **key** is the cipher itself, written down. It carries the character set and
the alphabet, derives nothing from a random number generator, and therefore
keeps working across Go versions, across programs and across languages:

```go
c := crogram.NewWithSeed(42)
key := c.Key()                     // "crogram1:61626364...:69323742..." (abridged)

same, err := crogram.ParseKey(key) // the exact same cipher
```

Prefer a key whenever the mapping has to outlive the process. Treat it as the
answer sheet: whoever holds it reads every message the cipher was used on.
`ParseKey` validates what it is given, so a corrupted or hostile key is
rejected (`ErrInvalidKey`, `ErrCharsetMismatch`, `ErrNotPermutation`,
`ErrDuplicateRune`) rather than producing an unusable cipher.

A **passphrase** is the friendly front door to a seed:

```go
c := crogram.NewFromPassphrase("correct horse battery staple")
```

SHA-256 of the passphrase becomes the seed. The cipher is then exactly as
guessable as the passphrase — and it is still a substitution cipher.

## Custom character sets

```go
c := crogram.NewCipher(
	crogram.WithCharset([]rune("abcdefghijklmnopqrstuvwxyz")),
	crogram.WithSeed(7),
)
```

Repeated runes are collapsed (first occurrence wins) and an empty set falls back
to `DefaultCharset` — and, if that is empty too, to a built-in set — so a cipher
always comes out usable and `NewCipher` cannot fail. `DefaultCharset` is a
variable, so a program may replace it globally: do it once, at start-up, before
any goroutine builds a cipher, because writing the variable itself is not safe
concurrently.

### Working in other languages

Any runes work as a character set — Greek, Cyrillic, Arabic, Hebrew, Japanese,
emoji — and the cipher treats them exactly as it treats ASCII. Two consequences
of Unicode are worth knowing, though; both come from the text rather than from
this package.

- **The default character set is ASCII.** In a language written with accents,
  those letters are outside the set and pass through *in clear*: a Portuguese
  `"ação"` keeps its `ç` and `ã` readable, which hands a solver two letters for
  free. Give the cipher your language's alphabet instead, or simply the runes of
  the text itself — `WithCharset` deduplicates what you give it:

  ```go
  c := crogram.NewCipher(crogram.WithCharset([]rune("abcdefghijklmnopqrstuvwxyzçãáàâéêíóôõú")))
  c = crogram.NewCipher(crogram.WithCharset([]rune("こんにちは世界")))   // any script
  ```

- **Substitution is per rune, not per letter as the reader sees it.** A combining
  accent is a rune of its own: `"café"` spelled with `é` (U+00E9) and with `e`
  followed by U+0301 encode differently, and both still decode back to what went
  in. If the text can arrive in any normalisation form, normalise it yourself —
  `golang.org/x/text/unicode/norm` does that, at the price of the one dependency
  this module has so far avoided.

`SuggestedGuess` and `EnglishFrequency` are English-shaped, since letter order
differs by language; reassign `EnglishFrequency` to your own order before
leaning on the first guess.

Explicit mappings are available too, for hand-built puzzles:

```go
c, err := crogram.NewFromPairs('a', 'q', 'b', 'x', 'c', 'm')   // a→q, b→x, c→m
c, err = crogram.NewFromMapping(map[rune]rune{'a': 'q'})
```

A replacement need not belong to the character set, which is what makes
substitutions such as `a` → `1` possible.

## Puzzles and solving

```go
puzzle := crogram.NewPuzzle("The quick brown fox", crogram.WithSeed(11))

fmt.Println(puzzle)        // the coded text a solver sees
fmt.Println(puzzle.Key())  // the answer sheet
```

For a solver working it out by hand, the package carries both halves of the
job — the analysis and the bookkeeping:

```go
profile := crogram.Analyse(puzzle.Ciphertext)
profile.Top(5)                                // commonest coded runes
profile.Rate('x')                             // how much of the text they are

puzzle.Hint(rng)                              // one true substitution
puzzle.Cipher.DecodeWith(text, knownLetters)  // apply guessed letters
puzzle.Cipher.SuggestedGuess(text)            // rank by English letter order
```

`SuggestedGuess` pairs the commonest coded runes with `EnglishFrequency`
(`etaoin…`). It is a starting point, not an answer: letter order only holds for
long, ordinary prose, and the package makes no attempt to check that the guess
means anything.

## Command line

```bash
go install github.com/rodrigocalmd/go-crogram/cmd/crogram@latest

echo "meet me at midnight" | crogram encode -seed 42
crogram decode -key "$KEY" "coded text"
crogram key -passphrase "correct horse battery staple"
crogram puzzle -seed 7 "The quick brown fox"
```

Flags come before the text — `crogram decode -seed 42 "coded text"` — and text
that itself starts with a dash has to be separated by `--`.

`encode` and `puzzle` print the key to standard error, so the coded text stays
clean for a pipe.

## API

| Symbol | Purpose |
|---|---|
| `New(seed ...int64)` | a cipher: random, or reproducible from a seed |
| `NewCipher(opts ...Option)` | a cipher shaped by options |
| `NewWithSeed(seed, opts…)` | a reproducible cipher |
| `NewFromPassphrase(passphrase, opts…)` | a reproducible cipher from a phrase |
| `NewFromPairs(pairs…)` / `NewFromMapping(m)` | explicit substitutions |
| `ParseKey(key)` / `MustParseKey(key)` | rebuild a cipher from its key |
| `WithSeed` / `WithCharset` / `WithDerangement` | options |
| `Cipher.Encode` / `Cipher.Decode` | substitute in either direction |
| `Cipher.DecodeWith(text, known)` | decode with guessed letters |
| `Cipher.Key` / `Cipher.String` | the answer sheet / a safe description |
| `Cipher.Pairs` / `Cipher.Hint` | every couple / one at random |
| `Cipher.Charset` / `Cipher.Alphabet` / `Cipher.Mapping` | copies of the tables |
| `Cipher.Replacement` / `Cipher.Original` | one rune's substitution, either way |
| `Cipher.Seed` / `Cipher.IsDeranged` / `Cipher.Len` | what this cipher is |
| `Analyse(text)` | letter-frequency profile |
| `Frequency.Counts` / `Rate` / `Sorted` / `Top` | the profile and its views |
| `Puzzle`, `NewPuzzle`, `NewPuzzleWith` | a challenge, its answer, `Hint` and `Solved` |

Errors are wrapped and testable with `errors.Is`: `ErrInvalidKey`,
`ErrCharsetMismatch`, `ErrNotPermutation`, `ErrDuplicateRune`,
`ErrUnpairedRune`.

## Concurrency

A `Cipher` is immutable once built — its tables are written only by the
constructors — so one `*Cipher` may be shared by any number of goroutines
calling `Encode`, `Decode`, `DecodeWith`, `Key` and the accessors at once.
`TestConcurrentUse` holds that promise under `-race`.

## Testing

```bash
make test     # go test ./...
make race     # go test -race ./...
make fuzz     # 30s of property-based round-trip fuzzing
make cover    # coverage summary
make fmt vet  # gofmt -w . and go vet ./...
```

The suite asserts properties rather than golden values: every charset rune is
replaced, the mapping is a bijection, the default cipher is a derangement,
seeds and passphrases reproduce, keys round-trip, bad keys are rejected, and
`Decode(Encode(x)) == x` holds for arbitrary input under the fuzzer.

CI runs `gofmt`, `go vet`, `go build` and `go test -race` on both the version
`go.mod` declares and the newest Go, fuzzes on the newest one, and lints with
`golangci-lint` at a pinned version using its default rules.

## Contributing

Issues and pull requests are welcome. Please run `make all` before opening a
pull request — it formats, vets and tests — and keep the tests property-based.

## License

MIT. See [LICENSE](LICENSE).
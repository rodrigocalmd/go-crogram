# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project aims
to follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

A broad revision of the package: the API grew options, portable keys, solver
helpers and a puzzle front end, and the documentation now says plainly what a
substitution cipher is and is not for.

### Added

- **Options.** `Option`, `WithSeed`, `WithCharset` and `WithDerangement`, so
  a cipher can be shaped without a new constructor per combination.
- **More constructors.** `NewCipher`, `NewWithSeed`, `NewFromPassphrase` (SHA-256 of a
  phrase becomes the seed), `NewFromPairs` and `NewFromMapping` for explicit
  substitutions.
- **Portable keys.** `Cipher.Key`, `ParseKey` and `MustParseKey`. A key carries
  the character set and the alphabet, so it rebuilds the cipher across Go
  versions, programs and languages.
- **Typed errors.** `ErrInvalidKey`, `ErrCharsetMismatch`, `ErrNotPermutation`,
  `ErrDuplicateRune` and `ErrUnpairedRune`, wrapped with context for
  `errors.Is`.
- **Cipher accessors.** `Charset`, `Alphabet`, `Mapping`, `Pairs`, `Hint`,
  `Replacement`, `Original`, `Seed`, `IsDeranged`, `Len` and `String`.
- **Solver helpers.** `Analyse` and the `Frequency` type (`Counts`, `Rate`,
  `Sorted`, `Top`), `EnglishFrequency`, `Cipher.SuggestedGuess` and
  `Cipher.DecodeWith` for decoding with guessed letters.
- **Puzzles.** `Puzzle`, `NewPuzzle` and `NewPuzzleWith`, with `Hint` and
  `Solved`.
- **Command line.** `cmd/crogram` with `encode`, `decode`, `key` and `puzzle`
  subcommands; `encode` and `puzzle` print the key to standard error.
- **Tests.** A property-based unit suite (bijection, derangement,
  reproducibility, key round-trip, bad keys, concurrency), fuzz targets
  (`FuzzRoundTrip`, `FuzzKeyRoundTrip`, `FuzzParseKeyNeverPanics`), executable
  examples for godoc, and a benchmark-free `make` workflow.
- **CI.** GitHub Actions running `gofmt`, `go vet`, `go build`,
  `go test -race`, a short fuzz run and `golangci-lint`.

### Changed

- **The generator is `math/rand/v2`.** A seeded cipher now draws its alphabet
  from PCG instead of `math/rand`'s LFSR, so **every seed maps to a different
  substitution than it did before**: a puzzle stored as a seed means something
  else now. Keys are unaffected — a key still rebuilds exactly the cipher it
  came from, and that is the form to keep. The same seed still replays through
  `NewWithSeed(Seed())`, and an unseeded cipher now draws from `math/rand/v2`'s
  global source instead of reading the clock.
- **`Hint` takes a `math/rand/v2` source.** `Cipher.Hint` and `Puzzle.Hint` now
  want a `*rand.Rand` from `math/rand/v2`: pass `rand.New(rand.NewPCG(1, 2))`
  for a reproducible draw, or `nil` to use the package's own source.
- **Requires Go 1.24.** The package uses nothing newer than Go 1.22
  (`math/rand/v2`, `slices`, `cmp`), so the floor is a choice rather than a
  need: recent enough to be widely installed, without pinning users to the
  newest release. CI tests that declared floor and the newest release, so the
  floor is real rather than aspirational.
- **Standard library use modernised.** `slices.Clone`, `slices.Sort`,
  `slices.SortFunc` and `cmp.Or` replace hand-rolled copies and `sort.Slice`,
  and the ranking rule is now one shared comparator.
- **CI.** Builds on the declared minimum and the newest Go, pins the linter
  version instead of taking `latest`, and uses current action releases.
- **`New` keeps its signature.** `New()` and `New(seed)` behave as before, so
  existing callers keep compiling; `NewCipher(opts ...Option)` is the new
  options-based constructor.
- **Derangement is on by default.** No rune stands in for itself, which is what
  a cryptogram wants. Pass `WithDerangement(false)` for the previous behaviour.
- **Documentation rewritten.** The README states outright that this is not
  cryptography, and explains when to use a seed and when to use a key.

### Fixed

- **The documented example outputs were impossible.** The README, `doc.go` and
  `Cipher.Encode` showed `"Hello World 123!"` encoding to `"Xq22l Dlee2 456!"`,
  which maps `l` to two different runes at once and so cannot come out of a
  cipher whose mapping is a bijection. The literals are now the values seed 42
  really produces, with a note that they follow `math/rand`'s stream.
- **`README` custom-charset example did not compile.** It passed options to
  `New`, which takes seeds; it now calls `NewCipher`.
- **`README` claimed Go 1.18 while `go.mod` declares 1.24.** The text now says
  what the module requires, and notes that only the `go` directive stands
  between the source and an older toolchain.
- **`IsDeranged` reported the option, not the mapping.** A one-rune character
  set cannot be deranged, and now says so instead of claiming otherwise.
- **`Seed` and `String` misdescribed key-built ciphers.** `Seed` promised that
  `NewWithSeed(Seed())` rebuilds the cipher, which is false for a cipher parsed
  from a key or built from explicit pairs; both now say where the cipher came
  from, and `String` reports `key`, `pairs` or `mapping` instead of `random`.
- **`ParseKey` accepted keys that were not valid UTF-8**, silently rebuilding a
  lossy cipher whose own `Key` no longer matched the key it came from. Those are
  rejected with `ErrInvalidKey` now.
- **An emptied `DefaultCharset` produced a cipher whose own key could not be
  parsed.** The fallback is a built-in set, so `New` always yields a usable
  cipher.
- **`cmd/crogram decode` required an origin.** It used to decode with a random
  cipher when given no `-key`, `-seed` or `-passphrase`, printing nonsense and
  exiting 0; it now fails with an explanation. `-charset` and `-derangement`
  are reported as ignored when `-key` is given, and `key` rejects a text
  argument instead of dropping it.
- **A flag after the text is now reported, not swallowed.** `crogram encode
  "text" -seed 42` used to encode the literal string `text -seed 42` with a
  cipher drawn at random and exit 0; `decode` answered with a message about
  missing flags. Both now name the flag that arrived too late and show the
  command's own form. The flag package stops parsing at the first positional
  argument, so the check lives in the command and text that looks like a flag
  still goes through an explicit `--`.
- **`LICENSE` added** (MIT), with the copyright holder matching the module path.

- **Seeds are recoverable.** `Cipher.Seed` reports the seed a cipher was built
  from, so an unseeded cipher can be replayed with `NewWithSeed`; the seed is
  resolved once and reused, so the value reported is the value used.
- **`Encode` and `Decode` allocate once**, sizing the buffer from the input
  instead of growing a slice per rune.
- **Derangement is drawn as a single-cycle permutation** (Sattolo's algorithm),
  which is a derangement by construction and needs no rejection sampling.
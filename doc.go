// Package crogram generates substitution ciphers ("cryptograms") over a
// configurable character set, and encodes and decodes text with them.
//
// A Cipher is built from a character set and a permutation of it. Encode
// replaces every rune of the set with its replacement and Decode maps it back,
// so encoding is one map lookup per rune. Runes outside the set — spaces,
// punctuation, newlines, or any rune from an alphabet the cipher does not
// cover — are passed through untouched, which keeps word shapes and layout
// intact. Invalid UTF-8 is passed through byte for byte as well, so encoding
// and decoding are exact for any input.
//
//	c := crogram.New()                        // a fresh cipher every run
//	encoded := c.Encode("Hello World 123!")   // coded text, different every run
//	c.Decode(encoded)                         // "Hello World 123!" again
//
// # This is not cryptography
//
// crogram is for puzzles, games, teaching and light obfuscation — not for
// confidentiality. A substitution cipher leaks letter frequency, word lengths
// and repeated words, and it falls to a few lines of frequency analysis or to
// a person with a pencil. Never use it to protect secrets, credentials or
// personal data; reach for crypto/aes or crypto/chacha20poly1305 and a real
// key-derivation function instead.
//
// # Determinism and portability
//
// A cipher built with WithSeed or NewFromPassphrase is reproducible: the same
// seed yields the same substitution, so a ciphertext can be recreated later
// without storing the mapping. The alphabet is drawn from math/rand/v2's PCG
// source, whose output for a given seed the standard library pins with a
// regression test — a strong habit of the standard library rather than a
// documented contract, which is the reason a key, and not a seed, is the form
// to keep.
//
// For something that survives across Go versions, across programs and across
// languages, use Key and ParseKey. A key carries the whole character set and
// the whole alphabet, and derives nothing:
//
//	c := crogram.NewWithSeed(42)
//	key := c.Key()                        // "crogram1:61626364...:69323742..." (abridged)
//	same, err := crogram.ParseKey(key)    // the exact same cipher
//
// # Custom character sets
//
// New defaults to DefaultCharset: the 26 lowercase letters, the 26 uppercase
// letters and the 10 digits. WithCharset replaces it. Repeated runes are
// collapsed (the first occurrence wins) and an empty set falls back to the
// default — or, if that was emptied too, to a built-in set — so New cannot fail
// and never panics.
//
// A set need not be ASCII: any runes work, and a script whose letters take
// three bytes costs nothing extra. Runes left out of the set pass through
// unchanged, which is worth remembering for a language written with accents —
// they are not in the default set, so they stay readable until the caller puts
// them in the charset.
//
// # Concurrency
//
// A Cipher is immutable once built: its lookup tables are written only by the
// constructors and read everywhere else. A single *Cipher may therefore be
// shared by any number of goroutines calling Encode, Decode, DecodeWith, Key
// and the accessors at the same time.
package crogram

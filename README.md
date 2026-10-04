# go-crogram

> [!WARNING]
> **This code is intended for entertainment and educational purposes. It is not a cryptographically secure encryption method.**
> Do not use it to protect passwords, personal data or any sensitive information.

A simple and flexible cryptogram generator package for Go.

This package allows you to create substitution ciphers for a character set that includes lowercase letters, uppercase letters, and numbers. You can generate a random cipher for one-time use or use a specific seed to create a reproducible cipher, allowing you to encode and decode messages across different sessions.

### Features

- Encodes and decodes text using a substitution cipher.
- Supports lowercase letters (`a-z`), uppercase letters (`A-Z`), and numbers (`0-9`).
- Characters not in the set (like spaces and punctuation) are preserved.
- Ciphers can be randomly generated or created from a specific seed for reproducibility.
- Efficient map-based implementation for fast lookups.

### Quickstart

Here's a basic example of how to use the package to generate a random cipher.

```go
package main

import (
	"fmt"

	"github.com/rodrigocalmd/go-crogram"
)

func main() {
	// Create a new random cipher
	randomCipher := crogram.New()

	originalText := "Hello World 123!"
	encodedText := randomCipher.Encode(originalText)
	fmt.Println("Encoded:", encodedText)

	decodedText := randomCipher.Decode(encodedText)
	fmt.Println("Decoded:", decodedText)
}
```
**Example Output**:
```
Encoded: j6FFu QuhFs 854!
Decoded: Hello World 123!
```
*(Note: Your output will be different due to the random nature of the cipher.)*

### Reproducible Ciphers with Seeds

If you need to decode a message in a later session, you must use the same cipher. You can achieve this by providing a seed when creating the cipher.

```go
package main

import (
	"fmt"

	"github.com/rodrigocalmd/go-crogram"
)

func main() {
	// Use a specific seed to create a reproducible cipher
	seed := int64(42)
	seededCipher := crogram.New(seed)

	originalText := "consistency matters"
	encodedText := seededCipher.Encode(originalText)
	fmt.Println("Encoded:", encodedText)

	// You can create another cipher with the same seed to decode the message
	anotherCipher := crogram.New(seed)
	decodedText := anotherCipher.Decode(encodedText)
	fmt.Println("Decoded:", decodedText)
}
```
**Output**:
```
Encoded: 9ZhB1BQSh9U ktQQS6B
Decoded: consistency matters
```

### Other languages (custom charsets)

`New` encodes `a-z`, `A-Z` and `0-9`; everything else is left as is. To encode other characters, use `NewWithCharset`:

```go
c, err := crogram.NewWithCharset(crogram.PortugueseCharset, 42)
// err is non-nil if the charset is empty, has fewer than 2 characters,
// repeats a character or is not valid UTF-8.
fmt.Println(c.Encode("Ação e emoção"))
```

Ready-made sets: `PortugueseCharset`, `SpanishCharset`, `RussianCharset`, or pass any string of distinct characters (kana, Greek, emoji...). Only characters in the set are encoded: with `DefaultCharset`, the `ç` in `ação` stays visible. The charset and the seed must both be known to decode. `NewWithCharset(DefaultCharset, seed)` is identical to `New(seed)`.

### Portable keys

A seed depends on Go's random number generator. A key does not: it is the 62-character substitution itself, so it keeps working across versions.

```go
key := crogram.New(42).Key()
c, err := crogram.FromKey(key) // error if the key is not a valid permutation
```

For a custom charset the key carries the charset too (`c<n>:<charset><substitutes>`), so `FromKey` alone is enough to decode.

### Streaming

```go
w := c.EncodeWriter(dst) // also: DecodeWriter, EncodeReader, DecodeReader
io.Copy(w, src)
w.Close() // flushes a character cut in half at the end; does not close dst
```

### Install and uninstall

```sh
# Install (needs Go; the binary goes to $(go env GOPATH)/bin, make sure it is in your PATH)
go install github.com/rodrigocalmd/go-crogram/cmd/crogram@latest

# Uninstall
rm "$(go env GOBIN)/crogram" 2>/dev/null || rm "$(go env GOPATH)/bin/crogram"
```

From a clone of the repository you can use `make install` and `make uninstall` (see `make help`).

### Command line

```sh

crogram encode "Hello World"          # prints the seed to stderr
crogram -s 42 encode "secret message"
echo "coded text" | crogram decode -s 42
crogram -c pt -s 42 encode "Ação"     # other languages: -c pt | es | ru | "your characters"
crogram key -s 42                      # print the portable key
crogram decode -k "$(crogram key -s 42)" -f in.txt -o out.txt
```

`-s/--seed` and `-k/--key` may be placed before or after the command. `decode` requires one of them. Input from `-f` or stdin is streamed.

`-c/--charset` takes a preset (`default`, `pt`, `es`, `ru`) or a list of characters. Decoding needs the same `-c` and `-s` (a key from `crogram key -c ...` already contains the charset, so `-k` alone is enough).

Other options: `-q/--quiet` (don't print the seed on encode), `-v/--version`, `-h/--help`. In a terminal, `encode` prints a hint such as `seed: 42   (decode with: crogram decode -s 42)`; when piped, it prints just the number.

### Development

```sh
make test              # go vet + go test -race -cover
make fuzz              # every fuzz target, 10s each (FUZZTIME=1m make fuzz)
make lint              # gofmt check
```

The fuzz targets check, with random input, that decoding always undoes encoding, that `FromKey` and `NewWithCharset` never panic, that streaming equals `Encode`/`Decode` for any chunk size (including characters cut in half), and that the CLI only exits with code 0, 1 or 2. If the fuzzer finds a failure, Go saves it in `testdata/fuzz/`; commit that file as a regression test.

### Examples

Everything above has a runnable example:

| What | Where | Run |
|---|---|---|
| Random cipher | `examples/simple.go` | `go run ./examples/simple.go` |
| Seeds (random, fixed, reading it back) | `examples/seed` | `go run ./examples/seed` |
| Keys (`Key`, `FromKey`, errors, custom charset) | `examples/key` | `go run ./examples/key` |
| Charsets (pt, es, ru, hiragana, emoji, errors) | `examples/charset` | `go run ./examples/charset` |
| Streaming (writers, readers, files, split characters) | `examples/streaming` | `go run ./examples/streaming` |
| Every CLI option | `examples/cli.sh` | `sh examples/cli.sh` |
| API examples checked by `go test` (shown in godoc) | `example_test.go` | `go test -run Example -v .` |

---
> [!WARNING]
> **This code is intended for entertainment and educational purposes. It is not a cryptographically secure encryption method.**

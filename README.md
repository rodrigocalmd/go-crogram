# go-crogram

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
	"crogram"
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
	"crogram"
)

func main() {
	// Use a specific seed to create a reproducible cipher
	seed := int64(42)
	seededCipher := crogram.New(seed)

	originalText := "consistency is key"
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
Encoded: 2v4sZstG42y Zs FGy
Decoded: consistency is key
```

---
*Note: This code is intended for entertainment and educational purposes. It is not a cryptographically secure encryption method.*

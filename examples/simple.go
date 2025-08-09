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

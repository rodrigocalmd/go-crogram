// Charsets: encode other languages and any set of characters.
package main

import (
	"fmt"

	"github.com/rodrigocalmd/go-crogram"
)

func show(name, charset, text string, seed int64) {
	c, err := crogram.NewWithCharset(charset, seed)
	if err != nil {
		fmt.Println(name, "error:", err)
		return
	}
	encoded := c.Encode(text)
	fmt.Printf("%-10s %q -> %q -> %q\n", name, text, encoded, c.Decode(encoded))
}

func main() {
	// Ready-made charsets.
	show("portuguese", crogram.PortugueseCharset, "Ação e emoção", 42)
	show("spanish", crogram.SpanishCharset, "¿Cómo está el niño?", 42)
	show("russian", crogram.RussianCharset, "Привет, мир!", 42)

	// Your own: any string of distinct characters.
	show("hiragana", "あいうえおかきくけこ", "あさ、こんにちは", 1)
	show("emoji", "😀😁😂🤣😃😄", "😀 hi 😄", 1)
	show("vowels", "aeiou", "banana", 1)

	// Characters outside the charset are NOT encoded: with the default
	// charset the "ç" and "ã" stay visible. Use a charset that has them.
	def := crogram.New(42)
	fmt.Println("default    ", def.Encode("ação")) // only "a" and "o" change

	// NewWithCharset(DefaultCharset, seed) is identical to New(seed).
	same, _ := crogram.NewWithCharset(crogram.DefaultCharset, 42)
	fmt.Println("same as New:", same.Encode("abc") == crogram.New(42).Encode("abc"))

	// Invalid charsets return an error instead of panicking.
	show("empty", "", "x", 1)
	show("one char", "a", "x", 1)
	show("repeated", "abca", "x", 1)
}

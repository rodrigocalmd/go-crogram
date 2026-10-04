package crogram_test

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rodrigocalmd/go-crogram"
)

// A cipher without a seed is random. Keep its seed (or key) if you need to
// decode the text in another run.
func ExampleNew_random() {
	c := crogram.New()
	seed, provided := c.Seed()
	_ = seed // save it somewhere to rebuild the cipher later

	encoded := c.Encode("Hello World 123!")
	fmt.Println(c.Decode(encoded))
	fmt.Println("seed provided by the caller:", provided)
	// Output:
	// Hello World 123!
	// seed provided by the caller: false
}

// The same seed always gives the same cipher, even in another program.
func ExampleNew_seed() {
	encoded := crogram.New(42).Encode("consistency matters")
	fmt.Println(encoded)

	// A different cipher object with the same seed decodes it.
	fmt.Println(crogram.New(42).Decode(encoded))
	// Output:
	// 9ZhB1BQSh9U ktQQS6B
	// consistency matters
}

// Only a-z, A-Z and 0-9 are encoded by New. Everything else is kept.
func ExampleCipher_Encode_passthrough() {
	c := crogram.New(42)
	fmt.Println(c.Encode("a, b! c? 日本 😀"))
	// Output: t, E! 9? 日本 😀
}

// Seed tells which seed built the cipher, and whether the caller gave it.
func ExampleCipher_Seed() {
	fmt.Println(crogram.New(7).Seed())
	_, provided := crogram.New().Seed()
	fmt.Println(provided)
	// Output:
	// 7 true
	// false
}

// A key is the substitution itself: it does not depend on the Go version
// or on the random number generator, so it is the safest thing to store.
func ExampleCipher_Key() {
	key := crogram.New(42).Key()
	fmt.Println(len(key))
	fmt.Println(key)
	// Output:
	// 62
	// tE9aSVr01zX5khZOj6BQJqodULD7CAwNgp8HbyiYnWRP3G2FfslMIu4TvcmKex
}

// FromKey rebuilds the cipher from a key. It rejects keys that are not a
// permutation of the charset.
func ExampleFromKey() {
	key := crogram.New(42).Key()

	c, err := crogram.FromKey(key)
	if err != nil {
		panic(err)
	}
	fmt.Println(c.Encode("consistency matters"))

	_, err = crogram.FromKey("too short")
	fmt.Println(err)
	// Output:
	// 9ZhB1BQSh9U ktQQS6B
	// crogram: key must have 62 characters, got 9
}

// The key of a custom charset contains the charset, so FromKey alone is
// enough to decode: no need to remember which charset was used.
func ExampleFromKey_customCharset() {
	original, _ := crogram.NewWithCharset("abcde", 1)
	key := original.Key()
	fmt.Println(key)

	c, _ := crogram.FromKey(key)
	fmt.Println(c.Decode(original.Encode("a bad cab")))
	// Output:
	// c5:abcdecabed
	// a bad cab
}

// Ready-made charsets cover accented letters and other alphabets.
func ExampleNewWithCharset_portuguese() {
	c, _ := crogram.NewWithCharset(crogram.PortugueseCharset, 42)
	encoded := c.Encode("Ação e emoção")
	fmt.Println(encoded)
	fmt.Println(c.Decode(encoded))
	// Output:
	// 4wsn 1 1pnwsn
	// Ação e emoção
}

func ExampleNewWithCharset_russian() {
	c, _ := crogram.NewWithCharset(crogram.RussianCharset, 42)
	encoded := c.Encode("Привет, мир!")
	fmt.Println(encoded)
	fmt.Println(c.Decode(encoded))
	// Output:
	// Ма4м7П, Ы4а!
	// Привет, мир!
}

// Any string of distinct characters is a charset: here only vowels and
// hiragana are encoded.
func ExampleNewWithCharset_custom() {
	c, _ := crogram.NewWithCharset("aeiouあいうえお", 3)
	encoded := c.Encode("banana あさ")
	fmt.Println(encoded)
	fmt.Println(c.Decode(encoded))
	// Output:
	// binini いさ
	// banana あさ
}

// NewWithCharset reports invalid charsets instead of panicking.
func ExampleNewWithCharset_invalid() {
	_, err := crogram.NewWithCharset("")
	fmt.Println(err)
	_, err = crogram.NewWithCharset("a")
	fmt.Println(err)
	_, err = crogram.NewWithCharset("abca")
	fmt.Println(err)
	// Output:
	// crogram: charset needs at least 2 characters, got 0
	// crogram: charset needs at least 2 characters, got 1
	// crogram: charset contains 'a' more than once
}

// With DefaultCharset the result is identical to New.
func ExampleNewWithCharset_default() {
	c, _ := crogram.NewWithCharset(crogram.DefaultCharset, 42)
	fmt.Println(c.Encode("consistency matters") == crogram.New(42).Encode("consistency matters"))
	// Output: true
}

// EncodeWriter encodes while writing, so large data never sits in memory.
// Always call Close at the end; it does not close the destination.
func ExampleCipher_EncodeWriter() {
	c := crogram.New(42)

	w := c.EncodeWriter(os.Stdout)
	io.WriteString(w, "consistency ")
	io.WriteString(w, "matters\n")
	w.Close()
	// Output: 9ZhB1BQSh9U ktQQS6B
}

func ExampleCipher_DecodeWriter() {
	c := crogram.New(42)

	w := c.DecodeWriter(os.Stdout)
	io.WriteString(w, "9ZhB1BQSh9U ktQQS6B\n")
	w.Close()
	// Output: consistency matters
}

// EncodeReader wraps any io.Reader: files, network connections, stdin...
func ExampleCipher_EncodeReader() {
	c := crogram.New(42)

	r := c.EncodeReader(strings.NewReader("consistency matters\n"))
	io.Copy(os.Stdout, r)
	// Output: 9ZhB1BQSh9U ktQQS6B
}

func ExampleCipher_DecodeReader() {
	c := crogram.New(42)

	r := c.DecodeReader(strings.NewReader("9ZhB1BQSh9U ktQQS6B\n"))
	io.Copy(os.Stdout, r)
	// Output: consistency matters
}

// Streaming also works with custom charsets, even when a multi-byte
// character such as "ç" arrives split across writes.
func ExampleCipher_EncodeWriter_customCharset() {
	c, _ := crogram.NewWithCharset(crogram.PortugueseCharset, 42)

	var encoded strings.Builder
	w := c.EncodeWriter(&encoded)
	data := []byte("Ação")
	w.Write(data[:3]) // "Aç" cut in the middle of the "ç"
	w.Write(data[3:])
	w.Close()

	fmt.Println(encoded.String() == c.Encode("Ação"))
	// Output: true
}

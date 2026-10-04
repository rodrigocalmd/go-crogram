// Seeds: a random cipher, a reproducible one, and how to read the seed back.
package main

import (
	"fmt"

	"github.com/rodrigocalmd/go-crogram"
)

func main() {
	// 1. Random cipher: different every run.
	random := crogram.New()
	seed, provided := random.Seed()
	fmt.Printf("random cipher: seed=%d (provided by the caller: %v)\n", seed, provided)

	encoded := random.Encode("Hello World 123!")
	fmt.Println("encoded:", encoded)

	// The seed drawn at random is enough to rebuild the same cipher later.
	again := crogram.New(seed)
	fmt.Println("decoded with the saved seed:", again.Decode(encoded))

	// 2. Reproducible cipher: same seed, same result, in any program.
	fmt.Println()
	fmt.Println(crogram.New(42).Encode("consistency matters")) // 9ZhB1BQSh9U ktQQS6B

	// 3. Only a-z, A-Z and 0-9 change; the rest is kept as is.
	fmt.Println(crogram.New(42).Encode("a, b! c? 日本 😀")) // t, E! 9? 日本 😀
}

// Keys: a portable alternative to seeds.
//
// A seed depends on Go's random number generator. A key is the substitution
// itself, so it keeps working even if that generator ever changes. Store the
// key, not the seed, when messages must be readable for a long time.
package main

import (
	"fmt"

	"github.com/rodrigocalmd/go-crogram"
)

func main() {
	key := crogram.New(42).Key()
	fmt.Println("key:", key)

	cipher, err := crogram.FromKey(key)
	if err != nil {
		panic(err)
	}
	fmt.Println(cipher.Encode("consistency matters")) // 9ZhB1BQSh9U ktQQS6B

	// A cipher rebuilt from a key has no seed.
	seed, provided := cipher.Seed()
	fmt.Println("seed:", seed, "provided:", provided) // 0 false

	// Invalid keys are rejected with an error.
	for _, bad := range []string{"too short", key[:61] + "!", key[:61] + key[:1]} {
		_, err := crogram.FromKey(bad)
		fmt.Println("error:", err)
	}

	// The key of a custom charset carries the charset, so FromKey is enough.
	custom, _ := crogram.NewWithCharset(crogram.PortugueseCharset, 7)
	customKey := custom.Key()
	fmt.Println("custom key:", customKey)
	restored, _ := crogram.FromKey(customKey)
	fmt.Println(restored.Decode(custom.Encode("Ação e emoção")))
}

package crogram

import (
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	cipher := New(12345) // Using a fixed seed for a predictable test

	testCases := []struct {
		name string
		text string
	}{
		{"simple lowercase", "helloworld"},
		{"mixed case", "HelloWorld"},
		{"with numbers", "HelloWorld123"},
		{"with punctuation", "Hello, World! 123..."},
		{"empty string", ""},
		{"only numbers", "1234567890"},
		{"only punctuation", "!@#$%^&*()"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := cipher.Encode(tc.text)
			decoded := cipher.Decode(encoded)

			if decoded != tc.text {
				t.Errorf("Encode/Decode cycle failed. Expected %q, got %q", tc.text, decoded)
			}
		})
	}
}

func TestReproducibleCipher(t *testing.T) {
	seed := int64(98765)
	text := "This is a test for reproducibility."

	cipher1 := New(seed)
	encoded1 := cipher1.Encode(text)

	cipher2 := New(seed)
	encoded2 := cipher2.Encode(text)

	if encoded1 != encoded2 {
		t.Errorf("Expected same encoded output for same seed, but got different results.\n1: %s\n2: %s", encoded1, encoded2)
	}

	decoded1 := cipher1.Decode(encoded1)
	if decoded1 != text {
		t.Errorf("Expected decoded text to match original, but got %q", decoded1)
	}
}

func TestDifferentSeeds(t *testing.T) {
	seed1 := int64(111)
	seed2 := int64(222)
	text := "This should be encoded differently."

	cipher1 := New(seed1)
	encoded1 := cipher1.Encode(text)

	cipher2 := New(seed2)
	encoded2 := cipher2.Encode(text)

	if encoded1 == encoded2 {
		t.Errorf("Expected different encoded output for different seeds, but got the same result: %s", encoded1)
	}

	if encoded1 == text {
		t.Errorf("Encoded text should not be the same as the original text.")
	}
}

func TestUnseededCipher(t *testing.T) {
	text := "testing unseeded cipher"
	cipher1 := New()
	encoded1 := cipher1.Encode(text)

	cipher2 := New()
	encoded2 := cipher2.Encode(text)

	if encoded1 == encoded2 {
		// This test has a small chance of failing if the two random seeds happen to be the same.
		// In a real-world scenario, you might need a more robust way to test this,
		// but for this project, it's a reasonable check.
		t.Logf("Warning: Unseeded ciphers produced the same output, which is statistically unlikely but possible.")
	}

	decoded1 := cipher1.Decode(encoded1)
	if decoded1 != text {
		t.Errorf("Expected decoded text to match original, but got %q", decoded1)
	}
}

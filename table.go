package crogram

// generateCharset creates the character set for the cryptogram.
// It includes lowercase letters, uppercase letters, and numbers.
func generateCharset() []rune {
	var charset []rune

	// Add lowercase letters 'a' through 'z'
	for r := 'a'; r <= 'z'; r++ {
		charset = append(charset, r)
	}

	// Add uppercase letters 'A' through 'Z'
	for r := 'A'; r <= 'Z'; r++ {
		charset = append(charset, r)
	}

	// Add numbers '0' through '9'
	for r := '0'; r <= '9'; r++ {
		charset = append(charset, r)
	}

	return charset
}

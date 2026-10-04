package crogram

// DefaultCharset is the character set used by New: lowercase letters,
// uppercase letters and digits, in this exact order. The order is part of
// the contract: together with the seed it determines the cipher, so it must
// never change.
const DefaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// defaultRunes is DefaultCharset as runes. It is only read, never modified.
var defaultRunes = []rune(DefaultCharset)

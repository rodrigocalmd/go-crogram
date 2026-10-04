package crogram

// Ready-made character sets for NewWithCharset. Characters that are not part
// of the chosen set are not encoded, so pick the set that matches the
// language of the text.
const (
	// PortugueseCharset is DefaultCharset plus Portuguese accented letters.
	PortugueseCharset = DefaultCharset + "áàâãçéêíóôõúü" + "ÁÀÂÃÇÉÊÍÓÔÕÚÜ"

	// SpanishCharset is DefaultCharset plus Spanish accented letters and ñ.
	SpanishCharset = DefaultCharset + "áéíóúüñ" + "ÁÉÍÓÚÜÑ"

	// RussianCharset is the Cyrillic alphabet (both cases) plus digits.
	RussianCharset = "абвгдеёжзийклмнопрстуфхцчшщъыьэюя" +
		"АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ" + "0123456789"
)

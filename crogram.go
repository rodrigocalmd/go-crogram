package crogram

import (
	"math/rand"
	"time"
)

type cryptogram struct {
	correct []TableASCII
	replace []int
}

type EncodeAndDecodeText struct {
	cryptogram cryptogram
}

func (c *EncodeAndDecodeText) Encode(text string) string {
	var encodeText []byte
	for _, v := range text {
		if v == 32 {
			encodeText = append(encodeText, byte(v))
			continue
		}
		for idx, elem := range c.cryptogram.correct {
			if rune(elem.Value) == v {
				char := rune(c.cryptogram.replace[idx])
				encodeText = append(encodeText, byte(char))
			}
		}
	}
	return string(encodeText)
}

func (d *EncodeAndDecodeText) Decode(text string) string {
	var decodeText []byte
	for _, v := range text {
		if v == 32 {
			decodeText = append(decodeText, byte(v))
			continue
		}
		for idx, elem := range d.cryptogram.replace {
			if rune(elem) == v {
				char := rune(d.cryptogram.correct[idx].Value)
				decodeText = append(decodeText, byte(char))
			}
		}
	}
	return string(decodeText)
}

func generateValues() cryptogram {
	var cryTable cryptogram
	cryTable.correct = Table
	rand.Seed(time.Now().UnixNano())
	lenT := len(cryTable.correct)
	copyCorrectValues := make([]int, lenT)
	for idx, elem := range cryTable.correct {
		copyCorrectValues[idx] = elem.Value
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(lenT, func(i, j int) {
		copyCorrectValues[i], copyCorrectValues[j] = copyCorrectValues[j], copyCorrectValues[i]
	})
	cryTable.replace = copyCorrectValues
	return cryTable
}

func Crogram() EncodeAndDecodeText {
	encDec := EncodeAndDecodeText{
		cryptogram: generateValues(),
	}
	return encDec
}

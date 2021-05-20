package main

import (
	"math/rand"
	"strconv"
	"time"
)

type Cryptogram struct {
	Correct []TableASCII `json:"correct"`
	Replace []int        `json:"replace"`
}

type EncodeAndDecodeText struct {
	TextEncoded string `json:"text_encoded"`
	TextDecode  string `json:"text_decode"`
	Cryptogram  Cryptogram
}

func (c *EncodeAndDecodeText) Encode(text string) {
	var encodeText []byte
	strASCII := strconv.QuoteToASCII(text)
	for _, v := range strASCII {
		if v == 32 {
			encodeText = append(encodeText, byte(v))
			continue
		}
		for idx, elem := range c.Cryptogram.Correct {
			if rune(elem.Value) == v {
				char := rune(c.Cryptogram.Replace[idx])
				encodeText = append(encodeText, byte(char))
			}
		}
	}
	c.TextEncoded = string(encodeText)
}

func (d *EncodeAndDecodeText) Decode(text string) {
	var decodeText []byte
	strASCII := strconv.QuoteToASCII(text)
	for _, v := range strASCII {
		if v == 32 {
			decodeText = append(decodeText, byte(v))
			continue
		}
		for idx, elem := range d.Cryptogram.Replace {
			if rune(elem) == v {
				char := rune(d.Cryptogram.Correct[idx].Value)
				decodeText = append(decodeText, byte(char))
			}
		}
	}
	d.TextDecode = string(decodeText)
}

func (c *EncodeAndDecodeText) GenerateValues() {
	var cryTable Cryptogram
	cryTable.Correct = Table
	rand.Seed(time.Now().UnixNano())
	lenT := len(cryTable.Correct)
	copyCorrectValues := make([]int, lenT)
	for idx, elem := range cryTable.Correct {
		copyCorrectValues[idx] = elem.Value
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(lenT, func(i, j int) {
		copyCorrectValues[i], copyCorrectValues[j] = copyCorrectValues[j], copyCorrectValues[i]
	})
	cryTable.Replace = copyCorrectValues
	c.Cryptogram = cryTable
}

func (e *EncodeAndDecodeText) EncodeCryptogram(text string) string {
	e.Encode(text)
	return e.TextEncoded
}

func (e *EncodeAndDecodeText) DecodeCryptogram(text string) string {
	e.Decode(text)
	return e.TextDecode
}

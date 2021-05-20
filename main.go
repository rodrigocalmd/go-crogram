package main

// func main() {
// 	init := Init()

// 	str := "hello world cry cry gram gram"
// 	enc := init.EncodeCryptogram(str)
// 	fmt.Println("Valor Cryptograma Encoded: " + enc)

// 	dec := init.DecodeCryptogram(enc)
// 	fmt.Println("Valor Cryptograma Decoded: " + dec)

// }

func Init() EncodeAndDecodeText {
	encDec := EncodeAndDecodeText{}
	encDec.GenerateValues()
	return encDec
}

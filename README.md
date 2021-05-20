# go-crogram


Cryptogram generator based on ASCII a-z table. (only lowercase letters without characters, for now)

## **Quickstart**

### **Example**:
```
package main

import (
	"fmt"

	"github.com/rodrigocalmd/go-crogram"
)

func main() {

	cryptogram := crogram.Crogram()

	strEncoded := cryptogram.Encode("hello world crogram crogram crogram")
	fmt.Println(strEncoded)

	strDecoded := cryptogram.Decode(strEncoded)
	fmt.Println(strDecoded)

}
```
**output**:
```
bqssg igzsj rzgnztl rzgnztl rzgnztl
hello world crogram crogram crogram
```
**OBS**: At each run, a new random table is generated. So after exiting the code it will not decode what was previously encoded.

#### **Note**: Code made for entertainment purposes only. Do not use for other purposes.

// Streaming: encode and decode io.Readers and io.Writers without loading
// everything in memory.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rodrigocalmd/go-crogram"
)

func main() {
	c := crogram.New(42)

	// 1. EncodeWriter: everything written is encoded and forwarded.
	// Always Close it when done (it does not close os.Stdout).
	fmt.Print("EncodeWriter: ")
	w := c.EncodeWriter(os.Stdout)
	io.WriteString(w, "consistency ")
	io.WriteString(w, "matters\n")
	w.Close()

	// 2. DecodeWriter: the opposite.
	fmt.Print("DecodeWriter: ")
	d := c.DecodeWriter(os.Stdout)
	io.WriteString(d, "9ZhB1BQSh9U ktQQS6B\n")
	d.Close()

	// 3. EncodeReader / DecodeReader wrap any io.Reader (file, network, stdin).
	fmt.Print("EncodeReader: ")
	io.Copy(os.Stdout, c.EncodeReader(strings.NewReader("consistency matters\n")))
	fmt.Print("DecodeReader: ")
	io.Copy(os.Stdout, c.DecodeReader(strings.NewReader("9ZhB1BQSh9U ktQQS6B\n")))

	// 4. Files: copy a file through the cipher.
	in, _ := os.CreateTemp("", "crogram-in-*.txt")
	defer os.Remove(in.Name())
	io.WriteString(in, "line one\nline two\n")
	in.Seek(0, io.SeekStart)

	out, _ := os.CreateTemp("", "crogram-out-*.txt")
	defer os.Remove(out.Name())

	ew := c.EncodeWriter(out)
	io.Copy(ew, in)
	ew.Close()
	out.Close()

	data, _ := os.ReadFile(out.Name())
	fmt.Printf("file encoded: %q\n", data)

	// 5. With a custom charset, a multi-byte character ("ç") may arrive cut
	// in half between two writes. Close flushes anything left at the end.
	pt, _ := crogram.NewWithCharset(crogram.PortugueseCharset, 42)
	var sb strings.Builder
	pw := pt.EncodeWriter(&sb)
	b := []byte("Ação")
	pw.Write(b[:3]) // "A" + first byte of "ç"
	pw.Write(b[3:]) // second byte of "ç" + "o"
	pw.Close()
	fmt.Println("split write equals Encode:", sb.String() == pt.Encode("Ação"))
}

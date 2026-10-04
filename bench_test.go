package crogram

import (
	"strings"
	"testing"
)

var sink string

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(int64(i))
	}
}

func BenchmarkEncode(b *testing.B) {
	c := New(1)
	text := strings.Repeat("Hello, World 123! ", 1<<16) // ~1.1 MB
	b.SetBytes(int64(len(text)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = c.Encode(text)
	}
}

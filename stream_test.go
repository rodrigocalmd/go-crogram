package crogram

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestStreams(t *testing.T) {
	c := New(5)
	text := strings.Repeat("Hello, Wörld 123!\n", 5000)
	want := c.Encode(text)

	var buf bytes.Buffer
	w := c.EncodeWriter(&buf)
	// Write in awkward chunks.
	for s := text; len(s) > 0; {
		n := 7
		if n > len(s) {
			n = len(s)
		}
		if _, err := io.WriteString(w, s[:n]); err != nil {
			t.Fatal(err)
		}
		s = s[n:]
	}
	if buf.String() != want {
		t.Error("EncodeWriter output differs from Encode output")
	}

	got, err := io.ReadAll(c.EncodeReader(iotest.OneByteReader(strings.NewReader(text))))
	if err != nil || string(got) != want {
		t.Errorf("EncodeReader output differs from Encode output (read error: %v)", err)
	}

	buf.Reset()
	io.Copy(c.DecodeWriter(&buf), strings.NewReader(want))
	if buf.String() != text {
		t.Error("DecodeWriter output differs from the original text")
	}
	got, _ = io.ReadAll(c.DecodeReader(strings.NewReader(want)))
	if string(got) != text {
		t.Error("DecodeReader output differs from the original text")
	}
}

func TestCharsetStreams(t *testing.T) {
	c := mustCharset(t, PortugueseCharset, 5)
	text := strings.Repeat("Coração, ação e emoção 😀 日本\n", 3000)
	want := c.Encode(text)

	// Writer, in chunks of 1 to 5 bytes: characters are split between writes.
	var buf bytes.Buffer
	w := c.EncodeWriter(&buf)
	for s, n := text, 1; len(s) > 0; n = n%5 + 1 {
		if n > len(s) {
			n = len(s)
		}
		if _, err := io.WriteString(w, s[:n]); err != nil {
			t.Fatal(err)
		}
		s = s[n:]
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if buf.String() != want {
		t.Error("EncodeWriter output differs from Encode output")
	}

	// Reader, one byte at a time.
	got, err := io.ReadAll(c.EncodeReader(iotest.OneByteReader(strings.NewReader(text))))
	if err != nil || string(got) != want {
		t.Errorf("EncodeReader output differs from Encode output (read error: %v)", err)
	}
	got, err = io.ReadAll(c.DecodeReader(iotest.OneByteReader(strings.NewReader(want))))
	if err != nil || string(got) != text {
		t.Errorf("DecodeReader output differs from the original text (read error: %v)", err)
	}

	// A truncated character at the end of the stream is flushed by Close, unchanged.
	buf.Reset()
	w = c.DecodeWriter(&buf)
	w.Write([]byte("a\xc3"))
	w.Close()
	if !strings.HasSuffix(buf.String(), "\xc3") {
		t.Errorf("incomplete trailing byte was lost: %q", buf.String())
	}
}

// chunkReader returns at most n bytes per Read, to split characters.
type chunkReader struct {
	r io.Reader
	n int
}

func (c *chunkReader) Read(p []byte) (int, error) {
	if len(p) > c.n {
		p = p[:c.n]
	}
	return c.r.Read(p)
}

// Whatever the chunk size, streaming must give exactly what Encode and
// Decode give for the whole text, including when a multi-byte character is
// cut in half and when the text is not valid UTF-8.
func FuzzStreamMatchesEncode(f *testing.F) {
	f.Add(int64(1), "Hello Wörld 123 ação 日本 😀", uint8(3), true)
	f.Add(int64(2), "plain ascii", uint8(1), false)
	f.Add(int64(3), "bad \xc3 bytes \xe3\x81", uint8(2), true)
	f.Fuzz(func(t *testing.T, seed int64, text string, chunk uint8, custom bool) {
		n := int(chunk)%16 + 1
		c := New(seed)
		if custom {
			c, _ = NewWithCharset(PortugueseCharset, seed)
		}
		wantEnc, wantDec := c.Encode(text), c.Decode(text)

		writeAll := func(w io.WriteCloser, s string) {
			for len(s) > 0 {
				k := min(n, len(s))
				if _, err := w.Write([]byte(s[:k])); err != nil {
					t.Fatal(err)
				}
				s = s[k:]
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
		}
		var enc, dec bytes.Buffer
		writeAll(c.EncodeWriter(&enc), text)
		writeAll(c.DecodeWriter(&dec), text)
		if enc.String() != wantEnc {
			t.Errorf("EncodeWriter (chunks of %d) gave %q, Encode gave %q", n, enc.String(), wantEnc)
		}
		if dec.String() != wantDec {
			t.Errorf("DecodeWriter (chunks of %d) gave %q, Decode gave %q", n, dec.String(), wantDec)
		}

		got, err := io.ReadAll(c.EncodeReader(&chunkReader{strings.NewReader(text), n}))
		if err != nil || string(got) != wantEnc {
			t.Errorf("EncodeReader (chunks of %d) gave %q (err %v), Encode gave %q", n, got, err, wantEnc)
		}
		got, err = io.ReadAll(c.DecodeReader(&chunkReader{strings.NewReader(text), n}))
		if err != nil || string(got) != wantDec {
			t.Errorf("DecodeReader (chunks of %d) gave %q (err %v), Decode gave %q", n, got, err, wantDec)
		}
	})
}

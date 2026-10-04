package crogram

import (
	"io"
	"unicode/utf8"
)

// streamer translates a byte stream. For byte-wise ciphers it is stateless;
// otherwise a multi-byte character split between two chunks is held back
// (carry) until its remaining bytes arrive.
type streamer struct {
	t        *tables
	byteWise bool
	carry    []byte
}

// process translates carry+src and appends the result to dst. Unless final is
// true, an incomplete trailing UTF-8 sequence is kept for the next call.
func (s *streamer) process(dst, src []byte, final bool) []byte {
	if s.byteWise {
		n := len(dst)
		dst = append(dst, src...)
		apply(dst[n:], src, s.t)
		return dst
	}
	data := src
	if len(s.carry) > 0 {
		data = append(s.carry, src...)
		s.carry = nil
	}
	end := len(data)
	if !final {
		end = completePrefix(data)
		if end < len(data) {
			s.carry = append([]byte(nil), data[end:]...)
		}
	}
	return appendTranslated(dst, data[:end], s.t)
}

// completePrefix returns the length of the longest prefix of b that does not
// end in the middle of a UTF-8 sequence.
func completePrefix(b []byte) int {
	for i := len(b) - 1; i >= 0 && i >= len(b)-utf8.UTFMax+1; i-- {
		if utf8.RuneStart(b[i]) {
			if utf8.FullRune(b[i:]) {
				return len(b)
			}
			return i
		}
	}
	return len(b)
}

type transformWriter struct {
	w   io.Writer
	s   streamer
	buf []byte
}

func (tw *transformWriter) Write(p []byte) (int, error) {
	tw.buf = tw.s.process(tw.buf[:0], p, false)
	if _, err := tw.w.Write(tw.buf); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close writes any bytes held back at the end of the stream. It does not
// close the underlying writer.
func (tw *transformWriter) Close() error {
	tw.buf = tw.s.process(tw.buf[:0], nil, true)
	if len(tw.buf) == 0 {
		return nil
	}
	_, err := tw.w.Write(tw.buf)
	return err
}

type transformReader struct {
	r    io.Reader
	s    streamer
	in   []byte
	out  []byte
	buf  []byte
	err  error
	done bool
}

func (tr *transformReader) Read(p []byte) (int, error) {
	if tr.s.byteWise {
		n, err := tr.r.Read(p)
		apply(p[:n], p[:n], tr.s.t)
		return n, err
	}
	if len(p) == 0 {
		return 0, nil
	}
	for len(tr.out) == 0 && !tr.done {
		if tr.in == nil {
			tr.in = make([]byte, 4096)
		}
		n, err := tr.r.Read(tr.in)
		if err != nil {
			tr.done, tr.err = true, err
		}
		tr.buf = tr.s.process(tr.buf[:0], tr.in[:n], err != nil)
		tr.out = tr.buf
	}
	if len(tr.out) > 0 {
		n := copy(p, tr.out)
		tr.out = tr.out[n:]
		return n, nil
	}
	return 0, tr.err
}

func (c *Cipher) streamer(t *tables) streamer {
	return streamer{t: t, byteWise: c.byteWise}
}

// EncodeWriter returns a writer that encodes everything written to it and
// forwards the result to w. Call Close when done: with a non byte-wise
// charset it flushes a character left incomplete at the end of the data.
// Close does not close w.
func (c *Cipher) EncodeWriter(w io.Writer) io.WriteCloser {
	return &transformWriter{w: w, s: c.streamer(&c.encode)}
}

// DecodeWriter is like EncodeWriter but decodes.
func (c *Cipher) DecodeWriter(w io.Writer) io.WriteCloser {
	return &transformWriter{w: w, s: c.streamer(&c.decode)}
}

// EncodeReader returns a reader that yields the encoded contents of r.
func (c *Cipher) EncodeReader(r io.Reader) io.Reader {
	return &transformReader{r: r, s: c.streamer(&c.encode)}
}

// DecodeReader returns a reader that yields the decoded contents of r.
func (c *Cipher) DecodeReader(r io.Reader) io.Reader {
	return &transformReader{r: r, s: c.streamer(&c.decode)}
}

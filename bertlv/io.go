package bertlv

import (
	"io"
)

type countReader struct {
	io.Reader
	Length *int64
}

func (c *countReader) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	*c.Length += int64(n)
	return
}

type countWriter struct {
	io.Writer
	Written *int64
}

func (w *countWriter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	*w.Written += int64(n)
	return
}

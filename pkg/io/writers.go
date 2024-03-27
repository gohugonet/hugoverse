package io

import (
	"io"
	"io/ioutil"
)

// As implemented by strings.Builder.
type FlexiWriter interface {
	io.Writer
	io.ByteWriter
	WriteString(s string) (int, error)
	WriteRune(r rune) (int, error)
}

// ToWriteCloser creates an io.WriteCloser from the given io.Writer.
// If it's not already, one will be created with a Close method that does nothing.
func ToWriteCloser(w io.Writer) io.WriteCloser {
	if rw, ok := w.(io.WriteCloser); ok {
		return rw
	}

	return struct {
		io.Writer
		io.Closer
	}{
		w,
		ioutil.NopCloser(nil),
	}
}

// NewMultiWriteCloser creates a new io.WriteCloser that duplicates its writes to all the
// provided writers.
func NewMultiWriteCloser(writeClosers ...io.WriteCloser) io.WriteCloser {
	writers := make([]io.Writer, len(writeClosers))
	for i, w := range writeClosers {
		writers[i] = w
	}
	return multiWriteCloser{Writer: io.MultiWriter(writers...), closers: writeClosers}
}

type multiWriteCloser struct {
	io.Writer
	closers []io.WriteCloser
}

func (m multiWriteCloser) Close() error {
	var err error
	for _, c := range m.closers {
		if closeErr := c.Close(); err != nil {
			err = closeErr
		}
	}
	return err
}

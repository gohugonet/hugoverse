package io

import (
	"bytes"
	"io"
	"strings"
)

// ReadSeekCloser is implemented by afero.File. We use this as the common type for
// content in Resource objects, even for strings.
type ReadSeekCloser interface {
	ReadSeeker
	io.Closer
}

// ReadSeeker wraps io.Reader and io.Seeker.
type ReadSeeker interface {
	io.Reader
	io.Seeker
}

// ReadSeekCloserProvider provides a ReadSeekCloser.
type ReadSeekCloserProvider interface {
	ReadSeekCloser() (ReadSeekCloser, error)
}

// NewReadSeekerNoOpCloserFromString uses strings.NewReader to create a new ReadSeekerNoOpCloser
// from the given string.
func NewReadSeekerNoOpCloserFromString(content string) ReadSeekerNoOpCloser {
	return ReadSeekerNoOpCloser{strings.NewReader(content)}
}

// NewReadSeekerNoOpCloserFromBytes uses strings.NewReader to create a new ReadSeekerNoOpCloser
// from the given bytes slice.
func NewReadSeekerNoOpCloserFromBytes(content []byte) ReadSeekerNoOpCloser {
	return ReadSeekerNoOpCloser{bytes.NewReader(content)}
}

// ReadSeekerNoOpCloser implements ReadSeekCloser by doing nothing in Close.
// TODO(bep) rename this and similar to ReadSeekerNopCloser, naming used in stdlib, which kind of makes sense.
type ReadSeekerNoOpCloser struct {
	ReadSeeker
}

// Close does nothing.
func (r ReadSeekerNoOpCloser) Close() error {
	return nil
}

// OpenReadSeekCloser allows setting some other way (than reading from a filesystem)
// to open or create a ReadSeekCloser.
type OpenReadSeekCloser func() (ReadSeekCloser, error)

// StringReader provides a way to read a string.
type StringReader interface {
	ReadString() string
}

// ReadString reads from the given reader and returns the content as a string.
func ReadString(r io.Reader) (string, error) {
	if sr, ok := r.(StringReader); ok {
		return sr.ReadString(), nil
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

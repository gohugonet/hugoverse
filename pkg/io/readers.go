package io

import (
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

// ReadSeekerNoOpCloser implements ReadSeekCloser by doing nothing in Close.
// TODO(bep) rename this and similar to ReadSeekerNopCloser, naming used in stdlib, which kind of makes sense.
type ReadSeekerNoOpCloser struct {
	ReadSeeker
}

// Close does nothing.
func (r ReadSeekerNoOpCloser) Close() error {
	return nil
}

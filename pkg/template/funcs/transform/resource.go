package transform

import (
	"github.com/mdfriday/hugoverse/pkg/io"
	"github.com/mdfriday/hugoverse/pkg/media"
)

// UnmarshableResource represents a Resource that can be unmarshaled to some other format.
type UnmarshableResource interface {
	ReadSeekCloserResource
	Identifier
}

// ReadSeekCloserResource is a Resource that supports loading its content.
type ReadSeekCloserResource interface {
	MediaType() media.Type
	io.ReadSeekCloserProvider
}

type Identifier interface {
	// Key is is mostly for internal use and should be considered opaque.
	// This value may change between Hugo versions.
	Key() string
}

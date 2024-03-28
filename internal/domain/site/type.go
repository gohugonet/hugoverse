package site

import (
	"bytes"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/spf13/afero"
	"io"
)

type Fs interface {
	Publish() afero.Fs
}

type ContentSpec interface {
	PreparePages() error
	RenderPages(func(kind string, sec []string, dir, name string, buf *bytes.Buffer) error) error
}

// Publisher publishes a result file.
type Publisher interface {
	Publish(d Descriptor) error
}

// Descriptor describes the needed publishing chain for an item.
type Descriptor struct {
	// The content to publish.
	Src io.Reader

	// The OutputFormat of this content.
	OutputFormat valueobject.Format

	// Where to publish this content. This is a filesystem-relative path.
	TargetPath string

	// If set, will replace all relative URLs with this one.
	AbsURLPath string
}

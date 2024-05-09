package site

import (
	"bytes"
	"github.com/gohugonet/hugoverse/pkg/output"
	"github.com/spf13/afero"
	"golang.org/x/text/collate"
	"io"
	"time"
)

type Site interface {
	URL
	Language
}

type Language interface {
	Location() *time.Location
	Collator() *collate.Collator
}

type URL interface {
	AbsURL(in string) string
	RelURL(in string) string
	URLize(uri string) string
}

type Config interface {
	URLConfig
	Languages() []LanguageConfig
}

type URLConfig interface {
	BaseUrl() string
}

type LanguageConfig interface {
	Name() string
	Code() string
}

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
	OutputFormat output.Format

	// Where to publish this content. This is a filesystem-relative path.
	TargetPath string

	// If set, will replace all relative URLs with this one.
	AbsURLPath string
}

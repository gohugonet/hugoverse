package site

import (
	"bytes"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/output"
	"github.com/spf13/afero"
	"golang.org/x/text/collate"
	"io"
	"time"
)

type Services interface {
	ContentService
	LanguageService
	FsService
	URLService
}

type LanguageService interface {
	LanguageKeys() []string
	GetLanguageIndex(lang string) (int, error)
}

type ContentService interface {
	WalkPages(langIndex int, walker contenthub.WalkFunc) error
}

type FsService interface {
	Publish() afero.Fs
}

type URLService interface {
	BaseUrl() string
}

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
}

type Template interface {
	MarkReady() error
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

package site

import (
	"bytes"
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/output"
	"github.com/spf13/afero"
	"golang.org/x/text/collate"
	"io"
	"time"
)

type Services interface {
	ContentService
	ResourceService
	LanguageService
	FsService
	URLService
	ConfigService
}

type ConfigService interface {
	ConfigParams() map[string]any
	SiteTitle() string
	Menus() map[string][]Menu
}

type Menu interface {
	Name() string
	URL() string
	Weight() int
}

type LanguageService interface {
	DefaultLanguage() string
	LanguageKeys() []string
	GetLanguageIndex(lang string) (int, error)
	GetLanguageName(lang string) string
}

type ContentService interface {
	WalkPages(langIndex int, walker contenthub.WalkFunc) error
	GetPageSources(page contenthub.Page) ([]contenthub.PageSource, error)
	GetPageFromPath(path string) (contenthub.Page, error)
	WalkTaxonomies(langIndex int, walker contenthub.WalkTaxonomyFunc) error
	GlobalPages() contenthub.Pages
}

type ResourceService interface {
	GetResourceWithOpener(pathname string, opener pio.OpenReadSeekCloser) (resources.Resource, error)
}

type FsService interface {
	Publish() afero.Fs
	WorkingDir() string
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

type Template interface {
	MarkReady() error
	LookupLayout(names []string) (template.Preparer, bool, error)
	ExecuteWithContext(ctx context.Context, t template.Preparer, wr io.Writer, data any) error
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

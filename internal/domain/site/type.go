package site

import (
	"bytes"
	"context"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/resources"
	"github.com/mdfriday/hugoverse/internal/domain/template"
	pio "github.com/mdfriday/hugoverse/pkg/io"
	"github.com/mdfriday/hugoverse/pkg/output"
	"github.com/spf13/afero"
	"golang.org/x/text/collate"
	"io"
	"time"
)

type Services interface {
	ContentService
	TranslationService
	ResourceService
	LanguageService
	FsService
	URLService
	ConfigService
	SitemapService
}

type SitemapService interface {
	ChangeFreq() string
	Priority() float64
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
	WalkTaxonomies(langIndex int, walker contenthub.WalkTaxonomyFunc) error
	GlobalPages(langIndex int) contenthub.Pages
	GlobalRegularPages() contenthub.Pages

	SearchPage(ctx context.Context, pages contenthub.Pages, page contenthub.Page) (contenthub.Pages, error)

	GetPageFromPath(langIndex int, path string) (contenthub.Page, error)
	GetPageRef(context contenthub.Page, ref string, home contenthub.Page) (contenthub.Page, error)
}

type TranslationService interface {
	Translate(ctx context.Context, lang string, translationID string, templateData any) string
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

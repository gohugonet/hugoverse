package contenthub

import (
	"bytes"
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
	"github.com/gohugonet/hugoverse/pkg/identity"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	goTmpl "html/template"
	"io"
)

type ContentHub interface {
	CollectPages() error
	PreparePages() error
	RenderPages(td TemplateDescriptor, cb func(info PageInfo) error) error

	RenderString(ctx context.Context, args ...any) (goTmpl.HTML, error)

	SetTemplateExecutor(exec Template)

	Title
}

type Title interface {
	CreateTitle(raw string) string
}

type TemplateDescriptor interface {
	Name() string
	Extension() string
}

type PageInfo interface {
	Kind() string
	Sections() []string
	Dir() string
	Name() string
	Buffer() *bytes.Buffer
}

//TODO remove unless we need to expose those kind for other domains

const (
	KindPage    = "page"
	KindHome    = "home"
	KindSection = "section"
)

var AllKindsInPages = []string{KindPage, KindHome, KindSection}

type Services interface {
	LangService
	FsService
	TaxonomyService
}

type FsService interface {
	NewFileMetaInfo(filename string) fs.FileMetaInfo

	LayoutFs() afero.Fs
	ContentFs() afero.Fs

	WalkContent(start string, cb fs.WalkCallback, conf fs.WalkwayConfig) error
}

type LangService interface {
	IsLanguageValid(lang string) bool
	GetSourceLang(source string) (string, bool)
	GetLanguageIndex(lang string) (int, error)
}

type Taxonomy interface {
	Singular() string // e.g. "category"
	Plural() string   // e.g. "categories"
}

type TaxonomyService interface {
	Views() []Taxonomy
}

type Template interface {
	ExecuteWithContext(ctx context.Context, tmpl template.Preparer, wr io.Writer, data any) error
	LookupVariants(name string) []template.Preparer
	LookupVariant(name string, variants template.Variants) (template.Preparer, bool, bool)
}

type BuildStateReseter interface {
	ResetBuildState()
}

type PageIdentity interface {
	identity.Identity

	Language() string
	LanguageIndex() int
}

type PageSource interface {
	PageIdentity() PageIdentity

	stale.Staler
	Language() string
	LanguageIndex() int

	Path() *paths.Path
	Opener() pio.OpenReadSeekCloser
}

type ContentNode interface {
	identity.IdentityProvider
	identity.ForEeachIdentityProvider
	stale.Marker
	BuildStateReseter

	Path() string
	IsContentNodeBranch() bool
}

type WeightedContentNode interface {
	ContentNode
	Weight() int
}

type ContentConvertProvider interface {
	GetContentConvertProvider(name string) ConverterProvider
}

type ConverterRegistry interface {
	Get(name string) ConverterProvider
}

// ConverterProvider creates converters.
type ConverterProvider interface {
	New(ctx markdown.DocumentContext) (Converter, error)
	Name() string
}

// ProviderProvider creates converter providers.
type ProviderProvider interface {
	New() (ConverterProvider, error)
}

// Converter wraps the Convert method that converts some markup into
// another format, e.g. Markdown to HTML.
type Converter interface {
	Convert(ctx markdown.RenderContext) (markdown.Result, error)
}

// ContentProvider provides the content related values for a Page.
type ContentProvider interface {
	Content() (any, error)
}

// FileProvider provides the source file.
type FileProvider interface {
	File() File
}

// File represents a source file.
// This is a temporary construct until we resolve page.Page conflicts.
type File interface {
	fileOverlap
	FileWithoutOverlap
}

// Temporary to solve duplicate/deprecated names in page.Page
type fileOverlap interface {
	// Path gets the relative path including file name and extension.
	// The directory is relative to the content root.
	RelPath() string

	// Section is first directory below the content root.
	// For page bundles in root, the Section will be empty.
	Section() string

	IsZero() bool
}

type FileWithoutOverlap interface {

	// Filename gets the full path and filename to the file.
	Filename() string

	// Dir gets the name of the directory that contains this file.
	// The directory is relative to the content root.
	Dir() string

	// Extension is an alias to Ext().
	// Deprecated: Use Ext instead.
	Extension() string

	// Ext gets the file extension, i.e "myblogpost.md" will return "md".
	Ext() string

	// LogicalName is filename and extension of the file.
	LogicalName() string

	// BaseFileName is a filename without extension.
	BaseFileName() string

	// TranslationBaseName is a filename with no extension,
	// not even the optional language extension part.
	TranslationBaseName() string

	// ContentBaseName is a either TranslationBaseName or name of containing folder
	// if file is a leaf bundle.
	ContentBaseName() string

	// UniqueID is the MD5 hash of the file's path and is for most practical applications,
	// Hugo content files being one of them, considered to be unique.
	UniqueID() string

	FileInfo() fs.FileMetaInfo
}

// Page is the core interface in Hugo.
type Page interface {
	RawContentProvider

	PageSource

	Kind() string
	Layouts() []string
	PageOutputs() ([]PageOutput, error)
}

type PageOutput interface {
	TargetFilePath() string
	TargetSubResourceDir() string
	TargetPrefix() string

	Content() (any, error)
}

type WalkFunc func(Page) error

// RawContentProvider provides the raw, unprocessed content of the page.
type RawContentProvider interface {
	// RawContent returns the raw, unprocessed content of the page excluding any front matter.
	RawContent() string
}

// PageWithoutContent is the Page without any of the content methods.
type PageWithoutContent interface {
	// FileProvider For pages backed by a file.
	FileProvider

	PageMetaProvider
}

// PageMetaProvider provides page metadata, typically provided via front matter.
type PageMetaProvider interface {

	// Kind The Page Kind. One of page, home, section, taxonomy, term.
	Kind() string

	// SectionsEntries Returns a slice of sections (directories if it's a file) to this
	// Page.
	SectionsEntries() []string

	// Layout The configured layout to use to render this page. Typically set in front matter.
	Layout() string

	// Type is a discriminator used to select layouts etc. It is typically set
	// in front matter, but will fall back to the root section.
	Type() string

	Lang() string
	Path() string
}

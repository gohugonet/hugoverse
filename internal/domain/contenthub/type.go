package contenthub

import (
	"bytes"
	"context"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/spf13/afero"
	"io"
)

type ContentHub interface {
	CollectPages() error
	PreparePages() error
	RenderPages(td TemplateDescriptor, cb func(info PageInfo) error) error
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

const (
	KindPage    = "page"
	KindHome    = "home"
	KindSection = "section"
)

var AllKindsInPages = []string{KindPage, KindHome, KindSection}

type Fs interface {
	LayoutFs() afero.Fs
	ContentFs() afero.Fs
}

type TemplateExecutor interface {
	ExecuteWithContext(ctx context.Context, tmpl template.Preparer, wr io.Writer, data any) error
	LookupLayout(d template.LayoutDescriptor) (template.Preparer, bool, error)
}

type ContentConvertProvider interface {
	GetContentConvertProvider(name string) ConverterProvider
}

type ConverterRegistry interface {
	Get(name string) ConverterProvider
}

// ConverterProvider creates converters.
type ConverterProvider interface {
	New(ctx DocumentContext) (Converter, error)
	Name() string
}

// ProviderProvider creates converter providers.
type ProviderProvider interface {
	New() (ConverterProvider, error)
}

// DocumentContext holds contextual information about the document to convert.
type DocumentContext struct {
	Document     any // May be nil. Usually a page.Page
	DocumentID   string
	DocumentName string
	Filename     string
}

// Converter wraps the Convert method that converts some markup into
// another format, e.g. Markdown to HTML.
type Converter interface {
	Convert(ctx RenderContext) (Result, error)
}

// RenderContext holds contextual information about the content to render.
type RenderContext struct {
	// Src is the content to render.
	Src []byte

	// Whether to render TableOfContents.
	RenderTOC bool

	// GerRenderer provides hook renderers on demand.
	//GetRenderer hooks.GetRendererFunc
}

// Result represents the minimum returned from Convert.
type Result interface {
	Bytes() []byte
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
	Path() string

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

	FileInfo() fsVO.FileMetaInfo
}

// Page is the core interface in Hugo.
type Page interface {
	ContentProvider
	PageWithoutContent
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
}

package markdown

import (
	"context"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/text"
	"github.com/gohugonet/hugoverse/pkg/types/hstring"
	"html/template"
	"io"
)

type Markdown interface {
	Render(rctx RenderContext, dctx DocumentContext) (Result, error)
}

// Result represents the minimum returned from Convert.
type Result interface {
	Bytes() []byte
}

type Highlighter interface {
	Highlight(code, lang string, opts any) (string, error)
	HighlightCodeBlock(ctx CodeblockContext, opts any) (HighlightResult, error)
	CodeBlockRenderer
	IsDefaultCodeBlockRendererProvider
}

type HighlightResult interface {
	Wrapped() template.HTML
	Inner() template.HTML
}

type AttributesOptionsSliceProvider interface {
	AttributesSlice() []Attribute
	OptionsSlice() []Attribute
}

type Attribute interface {
	Name() string
	Value() any
	ValueString() string
}

// RenderContext holds contextual information about the content to render.
type RenderContext struct {
	// Ctx is the context.Context for the current Page render.
	Ctx context.Context

	// Src is the content to render.
	Src []byte

	// Whether to render TableOfContents.
	RenderTOC bool

	// GerRenderer provides hook renderers on demand.
	GetRenderer GetRendererFunc
}

// DocumentContext holds contextual information about the document to convert.
type DocumentContext struct {
	Document     any // May be nil. Usually a page.Page
	DocumentID   string
	DocumentName string
	Filename     string
}

type RendererType int

const (
	LinkRendererType RendererType = iota + 1
	ImageRendererType
	HeadingRendererType
	CodeBlockRendererType
)

type GetRendererFunc func(t RendererType, id any) any

// ResultParse represents the minimum returned from Parse.
type ResultParse interface {
	Doc() any
	TableOfContentsProvider
}

type ContextData interface {
	RenderContext() RenderContext
	DocumentContext() DocumentContext
}

type TableOfContentsProvider interface {
	TableOfContents() TocFragments
}

type TocFragments interface {
	ToHTML(startLevel, stopLevel int, ordered bool) template.HTML
}

// Hooks

// LinkContext is the context passed to a link render hook.
type LinkContext interface {
	// The Page being rendered.
	Page() any

	// The link URL.
	Destination() string

	// The link title attribute.
	Title() string

	// The rendered (HTML) text.
	Text() hstring.RenderedString

	// The plain variant of Text.
	PlainText() string
}

type LinkRenderer interface {
	RenderLink(cctx context.Context, w io.Writer, ctx LinkContext) error
}

type AttributesProvider interface {
	// Attributes passed in from Markdown (e.g. { attrName1=attrValue1 attrName2="attr Value 2" }).
	Attributes() map[string]any
}

// HeadingContext contains accessors to all attributes that a HeadingRenderer
// can use to render a heading.
type HeadingContext interface {
	// Page is the page containing the heading.
	Page() any
	// Level is the level of the header (i.e. 1 for top-level, 2 for sub-level, etc.).
	Level() int
	// Anchor is the HTML id assigned to the heading.
	Anchor() string
	// Text is the rendered (HTML) heading text, excluding the heading marker.
	Text() hstring.RenderedString
	// PlainText is the unrendered version of Text.
	PlainText() string

	// Attributes (e.g. CSS classes)
	AttributesProvider
}

// HeadingRenderer describes a uniquely identifiable rendering hook.
type HeadingRenderer interface {
	// RenderHeading writes the rendered content to w using the data in w.
	RenderHeading(cctx context.Context, w io.Writer, ctx HeadingContext) error
}

type IsDefaultCodeBlockRendererProvider interface {
	IsDefaultCodeBlockRenderer() bool
}

// ElementPositionResolver provides a way to resolve the start Position
// of a markdown element in the original source document.
// This may be both slow and approximate, so should only be
// used for error logging.
type ElementPositionResolver interface {
	ResolvePosition(ctx any) text.Position
}

// CodeblockContext is the context passed to a code block render hook.
type CodeblockContext interface {
	AttributesProvider
	text.Positioner

	// Chroma highlighting processing options. This will only be filled if Type is a known Chroma Lexer.
	Options() map[string]any

	// The type of code block. This will be the programming language, e.g. bash, when doing code highlighting.
	Type() string

	// The text between the code fences.
	Inner() string

	// Zero-based ordinal for all code blocks in the current document.
	Ordinal() int

	// The owning Page.
	Page() any
}

type CodeBlockRenderer interface {
	RenderCodeblock(cctx context.Context, w pio.FlexiWriter, ctx CodeblockContext) error
}

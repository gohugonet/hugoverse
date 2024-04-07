package converter

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/markup/converter/hooks"
)

// DocumentContext holds contextual information about the document to convert.
type DocumentContext struct {
	Document     any // May be nil. Usually a page.Page
	DocumentID   string
	DocumentName string
	Filename     string
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
	GetRenderer hooks.GetRendererFunc
}

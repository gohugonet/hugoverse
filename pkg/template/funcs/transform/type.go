package transform

import (
	"context"
	"github.com/mdfriday/hugoverse/internal/domain/markdown"
	goTemplate "html/template"
)

type Markdown interface {
	RenderString(ctx context.Context, args ...any) (goTemplate.HTML, error)

	markdown.Highlighter
}

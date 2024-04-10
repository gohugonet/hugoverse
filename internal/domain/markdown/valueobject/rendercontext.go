package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
)

type RenderContextDataHolder struct {
	Rctx markdown.RenderContext
	Dctx markdown.DocumentContext
}

func (ctx *RenderContextDataHolder) RenderContext() markdown.RenderContext {
	return ctx.Rctx
}

func (ctx *RenderContextDataHolder) DocumentContext() markdown.DocumentContext {
	return ctx.Dctx
}

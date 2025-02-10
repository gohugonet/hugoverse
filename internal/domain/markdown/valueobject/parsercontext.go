package valueobject

import (
	"github.com/mdfriday/hugoverse/internal/domain/markdown"
	"github.com/yuin/goldmark/parser"
)

type ParserContext struct {
	parser.Context
}

func (p *ParserContext) TableOfContents() *Fragments {
	if v := p.Get(tocResultKey); v != nil {
		return v.(*Fragments)
	}
	return nil
}

func NewParserContext(rctx markdown.RenderContext) *ParserContext {
	ctx := parser.NewContext(parser.WithIDs(NewIDFactory(AutoHeadingIDTypeGitHub)))
	ctx.Set(tocEnableKey, rctx.RenderTOC)
	return &ParserContext{
		Context: ctx,
	}
}

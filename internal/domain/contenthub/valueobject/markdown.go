package valueobject

import (
	"bytes"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/markup/converter"
	"github.com/gohugonet/hugoverse/pkg/markup/gdm"
	"github.com/gohugonet/hugoverse/pkg/markup/gdm/render"
	"github.com/gohugonet/hugoverse/pkg/markup/tableofcontents"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// MDProvider is the package entry point.
var MDProvider contenthub.ProviderProvider = provide{}

type provide struct {
	name string
}

func (p provide) New() (contenthub.ConverterProvider, error) {
	return ConverterProvider{
		name: "markdown",
		create: func(ctx converter.DocumentContext) (contenthub.Converter, error) {
			builder := gdm.NewBuilder(gdm.DefaultConf)
			md := builder.Build()

			return &mdConverter{
				md:  md,
				ctx: ctx,
			}, nil
		},
	}, nil
}

type mdConverter struct {
	md  goldmark.Markdown
	ctx converter.DocumentContext
}

func (c *mdConverter) Convert(ctx converter.RenderContext) (result contenthub.Result, err error) {
	parseResult, err := c.Parse(ctx)
	if err != nil {
		return nil, err
	}

	renderResult, err := c.Render(ctx, parseResult.Doc())
	if err != nil {
		return nil, err
	}

	return converterResult{
		Result:                  renderResult,
		tableOfContentsProvider: parseResult,
	}, nil
}

func (c *mdConverter) Parse(ctx converter.RenderContext) (contenthub.ResultParse, error) {
	pctx := c.newParserContext(ctx)
	reader := text.NewReader(ctx.Src)

	doc := c.md.Parser().Parse(
		reader,
		parser.WithContext(pctx),
	)

	return parserResult{
		doc: doc,
		toc: pctx.TableOfContents(),
	}, nil
}

func (c *mdConverter) Render(ctx converter.RenderContext, doc any) (contenthub.Result, error) {
	n := doc.(ast.Node)
	buf := &render.BufWriter{Buffer: &bytes.Buffer{}}

	rcx := &render.RenderContextDataHolder{
		Rctx: ctx,
		Dctx: c.ctx,
	}

	w := &render.Context{
		BufWriter:   buf,
		ContextData: rcx,
	}

	if err := c.md.Renderer().Render(w, ctx.Src, n); err != nil {
		return nil, err
	}

	return renderResult{
		Result: buf,
	}, nil
}

type renderResult struct {
	contenthub.Result
}

var (
	tocResultKey = parser.NewContextKey()
	tocEnableKey = parser.NewContextKey()
)

func (c *mdConverter) newParserContext(rctx converter.RenderContext) *parserContext {
	ctx := parser.NewContext(parser.WithIDs(gdm.NewIDFactory(gdm.AutoHeadingIDTypeGitHub)))
	ctx.Set(tocEnableKey, rctx.RenderTOC)
	return &parserContext{
		Context: ctx,
	}
}

type parserContext struct {
	parser.Context
}

func (p *parserContext) TableOfContents() *tableofcontents.Fragments {
	if v := p.Get(tocResultKey); v != nil {
		return v.(*tableofcontents.Fragments)
	}
	return nil
}

type parserResult struct {
	doc any
	toc *tableofcontents.Fragments
}

func (p parserResult) Doc() any {
	return p.doc
}

func (p parserResult) TableOfContents() *tableofcontents.Fragments {
	return p.toc
}

type converterResult struct {
	contenthub.Result
	tableOfContentsProvider
}

type tableOfContentsProvider interface {
	TableOfContents() *tableofcontents.Fragments
}

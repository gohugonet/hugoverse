package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"github.com/gohugonet/hugoverse/internal/domain/markdown/factory"
)

// MDProvider is the package entry point.
var MDProvider contenthub.ProviderProvider = provide{}

type provide struct {
	name string
}

func (p provide) New() (contenthub.ConverterProvider, error) {
	return ConverterProvider{
		name: "markdown",
		create: func(ctx markdown.DocumentContext) (contenthub.Converter, error) {

			return &mdConverter{
				md:  factory.NewMarkdown(),
				ctx: ctx,
			}, nil
		},
	}, nil
}

type mdConverter struct {
	md  markdown.Markdown
	ctx markdown.DocumentContext
}

func (c *mdConverter) Convert(ctx markdown.RenderContext) (result contenthub.Result, err error) {
	return c.md.Render(ctx, c.ctx)
}

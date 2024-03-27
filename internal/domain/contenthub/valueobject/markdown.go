package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
)

// MDProvider is the package entry point.
var MDProvider contenthub.ProviderProvider = provide{}

type provide struct {
	name string
}

func (p provide) New() (contenthub.Provider, error) {
	//TODO, implement me with dddplayer/markdown
	// md := newMarkdown()

	return ConverterProvider{
		name: "markdown",
		create: func(ctx contenthub.DocumentContext) (contenthub.Converter, error) {
			return &mdConverter{}, nil
		},
	}, nil
}

type mdConverter struct {
	//md dddplayer.markdown
}

func (c *mdConverter) Convert(ctx contenthub.RenderContext) (result contenthub.Result, err error) {
	fmt.Println("markdown >>> ...", string(ctx.Src))

	return converterResult{bytes: ctx.Src}, nil
}

type converterResult struct {
	bytes []byte
}

func (c converterResult) Bytes() []byte {
	return c.bytes
}

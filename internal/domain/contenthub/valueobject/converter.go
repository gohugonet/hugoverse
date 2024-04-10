package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"strings"
)

type ConverterRegistry struct {
	// Maps name (md, markdown, goldmark etc.) to a converter provider.
	// Note that this is also used for aliasing, so the same converter
	// may be registered multiple times.
	// All names are lower case.
	Converters map[string]contenthub.ConverterProvider
}

func (r *ConverterRegistry) Get(name string) contenthub.ConverterProvider {
	return r.Converters[strings.ToLower(name)]
}

type ConverterProvider struct {
	name   string
	create func(ctx markdown.DocumentContext) (contenthub.Converter, error)
}

func (n ConverterProvider) New(ctx markdown.DocumentContext) (contenthub.Converter, error) {
	return n.create(ctx)
}

func (n ConverterProvider) Name() string {
	return n.name
}

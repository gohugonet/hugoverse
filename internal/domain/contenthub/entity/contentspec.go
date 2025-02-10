package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"strings"
)

type ContentSpec struct {
	Converters contenthub.ConverterRegistry
}

func (c *ContentSpec) ResolveMarkup(in string) string {
	in = strings.ToLower(in)
	switch in {
	case "md", "markdown", "mdown":
		return "markdown"
	case "html", "htm":
		return "html"
	default:
		if conv := c.Converters.Get(in); conv != nil {
			return conv.Name()
		}
	}
	return ""
}

func (c *ContentSpec) GetContentConvertProvider(name string) contenthub.ConverterProvider {
	return c.Converters.Get(name)
}

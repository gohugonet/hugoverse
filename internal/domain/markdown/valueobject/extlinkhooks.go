package valueobject

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

func NewLinkHooks(protocol string) goldmark.Extender {
	return &links{LinkifyProtocol: protocol}
}

type links struct {
	LinkifyProtocol string
}

// Extend implements goldmark.Extender.
func (e *links) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newLinkRenderer(e.LinkifyProtocol), 100),
	))
}

func newLinkRenderer(protocol string) renderer.NodeRenderer {
	r := &hookedRenderer{
		linkifyProtocol: []byte(protocol),
		Config: html.Config{
			Writer: html.DefaultWriter,
		},
	}
	return r
}

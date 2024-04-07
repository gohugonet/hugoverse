package gdm

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

func newLinks(cfg Config) goldmark.Extender {
	return &links{cfg: cfg}
}

type links struct {
	cfg Config
}

// Extend implements goldmark.Extender.
func (e *links) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newLinkRenderer(e.cfg), 100),
	))
}

func newLinkRenderer(cfg Config) renderer.NodeRenderer {
	r := &hookedRenderer{
		linkifyProtocol: []byte(cfg.Extensions.LinkifyProtocol),
		Config: html.Config{
			Writer: html.DefaultWriter,
		},
	}
	return r
}

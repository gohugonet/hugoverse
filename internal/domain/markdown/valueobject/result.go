package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
)

type Result struct {
	markdown.RenderResult
	markdown.TableOfContentsProvider
}

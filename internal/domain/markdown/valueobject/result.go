package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
)

type Result struct {
	markdown.Result
	markdown.TableOfContentsProvider
}

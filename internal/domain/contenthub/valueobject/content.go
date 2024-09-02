package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"html/template"
)

type ContentSummary struct {
	Content          template.HTML
	Summary          template.HTML
	SummaryTruncated bool
}

type ContentToC struct {
	tableOfContents     markdown.TocFragments
	tableOfContentsHTML template.HTML
}

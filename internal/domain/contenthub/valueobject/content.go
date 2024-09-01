package valueobject

import "html/template"

type ContentSummary struct {
	Content          template.HTML
	Summary          template.HTML
	SummaryTruncated bool
}

type ContentToC struct {
	// For Goldmark we split Parse and Render.
	astDoc any

	tableOfContents     *tableofcontents.Fragments
	tableOfContentsHTML template.HTML
}

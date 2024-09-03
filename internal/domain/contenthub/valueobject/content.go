package valueobject

import (
	"html/template"
)

type ContentSummary struct {
	Content             template.HTML
	Summary             template.HTML
	SummaryTruncated    bool
	TableOfContentsHTML template.HTML
}

func NewEmptyContentSummary() ContentSummary {
	return ContentSummary{
		Content:             "",
		Summary:             "",
		SummaryTruncated:    false,
		TableOfContentsHTML: "",
	}
}

// DefaultTocConfig is the default ToC configuration.
var DefaultTocConfig = TocConfig{
	StartLevel: 2,
	EndLevel:   3,
	Ordered:    false,
}

type TocConfig struct {
	// Heading start level to include in the table of contents, starting
	// at h1 (inclusive).
	// <docsmeta>{ "identifiers": ["h1"] }</docsmeta>
	StartLevel int

	// Heading end level, inclusive, to include in the table of contents.
	// Default is 3, a value of -1 will include everything.
	EndLevel int

	// Whether to produce a ordered list or not.
	Ordered bool
}

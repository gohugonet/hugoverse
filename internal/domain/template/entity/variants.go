package entity

import "github.com/gohugonet/hugoverse/pkg/output"

// TemplateVariants describes the possible variants of a template.
// All of these may be empty.
type TemplateVariants struct {
	Language     string
	OutputFormat output.Format
}

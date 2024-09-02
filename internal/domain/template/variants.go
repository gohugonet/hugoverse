package template

import "github.com/gohugonet/hugoverse/pkg/output"

// Variants describes the possible variants of a template.
// All of these may be empty.
type Variants struct {
	Language     string
	OutputFormat output.Format
}

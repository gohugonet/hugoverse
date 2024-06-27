package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/template"

type Shortcode struct {
	Name string

	Ordinal int

	Indentation string // indentation from source.

	Inner    []any // string or nested shortcode
	IsInline bool  // inline shortcode. Any inner will be a Go template.

	// If set, the rendered shortcode is sent as part of the surrounding content
	// to Goldmark and similar.
	// Before Hug0 0.55 we didn't send any shortcode output to the markup
	// renderer, and this flag told Hugo to process the {{ .Inner }} content
	// separately.
	// The old behavior can be had by starting your shortcode template with:
	//    {{ $_hugo_config := `{ "version": 1 }`}}
	DoMarkup bool

	IsClosing bool // whether a closing tag was provided

	Info   template.Info       // One of the output formats (arbitrary)
	Templs []template.Preparer // All output formats

	Params any // map or array

	Pos    int // the position in bytes in the source file
	Length int // the length in bytes in the source file

	// the placeholder in the source when passed to Goldmark etc.
	// This also identifies the rendered shortcode.
	Placeholder string
}

func (s Shortcode) NeedsInner() bool {
	return s.Info != nil && s.Info.ParseInfo().Inner()
}

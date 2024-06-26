package valueobject

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
}

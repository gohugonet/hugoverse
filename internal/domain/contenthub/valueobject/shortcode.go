package valueobject

import (
	"context"
	"github.com/mdfriday/hugoverse/internal/domain/template"
)

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

func (s Shortcode) InsertPlaceholder() bool {
	return !s.DoMarkup
}

// ShortcodeRenderer is typically used to delay rendering of inner shortcodes
// marked with placeholders in the content.
type ShortcodeRenderer interface {
	RenderShortcode(context.Context) ([]byte, bool, error)
	RenderShortcodeString(context.Context) (string, bool, error)
}

type ShortcodeRenderFunc func(context.Context) ([]byte, bool, error)

func (f ShortcodeRenderFunc) RenderShortcode(ctx context.Context) ([]byte, bool, error) {
	return f(ctx)
}

func (f ShortcodeRenderFunc) RenderShortcodeString(ctx context.Context) (string, bool, error) {
	b, has, err := f(ctx)
	return string(b), has, err
}

type prerenderedShortcode struct {
	s           string
	hasVariants bool
}

func (p prerenderedShortcode) RenderShortcode(context.Context) ([]byte, bool, error) {
	return []byte(p.s), p.hasVariants, nil
}

func (p prerenderedShortcode) RenderShortcodeString(context.Context) (string, bool, error) {
	return p.s, p.hasVariants, nil
}

func NewPrerenderedShortcode(s string, hasVariants bool) ShortcodeRenderer {
	return prerenderedShortcode{s: s, hasVariants: hasVariants}
}

var ZeroShortcode = prerenderedShortcode{}

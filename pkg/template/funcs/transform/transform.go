package transform

import (
	"bytes"
	"context"
	"encoding/xml"
	"github.com/mdfriday/hugoverse/pkg/cache/dynacache"
	"github.com/mdfriday/hugoverse/pkg/cache/stale"
	"github.com/mdfriday/hugoverse/pkg/helpers"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"html"
	"html/template"
	"strings"

	"github.com/spf13/cast"
)

// New returns a new instance of the transform-namespaced template functions.
func New(md Markdown) *Namespace {
	memCache := dynacache.New(dynacache.Options{Running: true, Log: loggers.NewDefault()})
	return &Namespace{
		cache: dynacache.GetOrCreatePartition[string, *stale.Value[any]](
			memCache,
			"/tmpl/transform",
			dynacache.OptionsPartition{Weight: 30, ClearWhen: dynacache.ClearOnChange},
		),
		md: md,
	}
}

// Namespace provides template functions for the "transform" namespace.
type Namespace struct {
	cache *dynacache.Partition[string, *stale.Value[any]]
	md    Markdown
}

// Emojify returns a copy of s with all emoji codes replaced with actual emojis.
//
// See http://www.emoji-cheat-sheet.com/
func (ns *Namespace) Emojify(s any) (template.HTML, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	return template.HTML(helpers.Emojify([]byte(ss))), nil
}

// Highlight returns a copy of s as an HTML string with syntax
// highlighting applied.
func (ns *Namespace) Highlight(s any, lang string, opts ...any) (template.HTML, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	var optsv any
	if len(opts) > 0 {
		optsv = opts[0]
	}

	highlighted, err := ns.md.Highlight(ss, lang, optsv)
	if err != nil {
		return "", err
	}
	return template.HTML(highlighted), nil
}

// HTMLEscape returns a copy of s with reserved HTML characters escaped.
func (ns *Namespace) HTMLEscape(s any) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	return html.EscapeString(ss), nil
}

// HTMLUnescape returns a copy of s with HTML escape requences converted to plain
// text.
func (ns *Namespace) HTMLUnescape(s any) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	return html.UnescapeString(ss), nil
}

// XMLEscape returns the given string, removing disallowed characters then
// escaping the result to its XML equivalent.
func (ns *Namespace) XMLEscape(s any) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	// https://www.w3.org/TR/xml/#NT-Char
	cleaned := strings.Map(func(r rune) rune {
		if r == 0x9 || r == 0xA || r == 0xD ||
			(r >= 0x20 && r <= 0xD7FF) ||
			(r >= 0xE000 && r <= 0xFFFD) ||
			(r >= 0x10000 && r <= 0x10FFFF) {
			return r
		}
		return -1
	}, ss)

	var buf bytes.Buffer
	err = xml.EscapeText(&buf, []byte(cleaned))
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Markdownify renders s from Markdown to HTML.
func (ns *Namespace) Markdownify(ctx context.Context, s any) (template.HTML, error) {
	return ns.md.RenderString(ctx, s)
}

// Plainify returns a copy of s with all HTML tags removed.
func (ns *Namespace) Plainify(s any) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	return helpers.StripHTML(ss), nil
}

// For internal use.
func (ns *Namespace) Reset() {
	ns.cache.Clear()
}

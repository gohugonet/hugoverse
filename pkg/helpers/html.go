package helpers

import (
	bp "github.com/mdfriday/hugoverse/pkg/bufferpool"
	htmltemplate "github.com/mdfriday/hugoverse/pkg/template/htmltemplate"
	"html/template"
	"strings"
	"unicode"
)

const hugoNewLinePlaceholder = "___hugonl_"

var stripHTMLReplacerPre = strings.NewReplacer("\n", " ", "</p>", hugoNewLinePlaceholder, "<br>", hugoNewLinePlaceholder, "<br />", hugoNewLinePlaceholder)

// StripHTML strips out all HTML tags in s.
func StripHTML(s string) string {
	// Shortcut strings with no tags in them
	if !strings.ContainsAny(s, "<>") {
		return s
	}

	pre := stripHTMLReplacerPre.Replace(s)
	preReplaced := pre != s

	s = htmltemplate.StripTags(pre)

	if preReplaced {
		s = strings.ReplaceAll(s, hugoNewLinePlaceholder, "\n")
	}

	var wasSpace bool
	b := bp.GetBuffer()
	defer bp.PutBuffer(b)
	for _, r := range s {
		isSpace := unicode.IsSpace(r)
		if !(isSpace && wasSpace) {
			b.WriteRune(r)
		}
		wasSpace = isSpace
	}

	if b.Len() > 0 {
		s = b.String()
	}

	return s
}

// BytesToHTML converts bytes to type template.HTML.
func BytesToHTML(b []byte) template.HTML {
	return template.HTML(string(b))
}

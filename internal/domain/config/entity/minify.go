package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/config/valueobject"
	"github.com/mdfriday/hugoverse/pkg/media"
	"github.com/tdewolff/minify/v2"
	"io"
)

type MinifyC struct {
	valueobject.MinifyConfig
}

func (m MinifyC) IsMinifyPublish() bool {
	return m.MinifyOutput
}

func (m MinifyC) GetMinifier(s string) minify.Minifier {
	switch {
	case s == "css" && !m.DisableCSS:
		return &m.Tdewolff.CSS
	case s == "js" && !m.DisableJS:
		return &m.Tdewolff.JS
	case s == "json" && !m.DisableJSON:
		return &m.Tdewolff.JSON
	case s == "svg" && !m.DisableSVG:
		return &m.Tdewolff.SVG
	case s == "xml" && !m.DisableXML:
		return &m.Tdewolff.XML
	case s == "html" && !m.DisableHTML:
		return &m.Tdewolff.HTML
	default:
		return noopMinifier{}
	}
}

func (m MinifyC) Minifiers(mediaTypes media.Types, cb func(media.Type, minify.Minifier)) {
	for _, suffix := range []string{"css", "js", "json", "svg", "xml", "html"} {
		types := mediaTypes.BySuffix(suffix)
		mi := m.GetMinifier(suffix)
		for _, t := range types {
			cb(t, mi)
		}
	}
}

// noopMinifier implements minify.Minifier [1], but doesn't minify content. This means
// that we can avoid missing minifiers for any MIME types in our minify.M, which
// causes minify to return errors, while still allowing minification to be
// disabled for specific types.
//
// [1]: https://pkg.go.dev/github.com/tdewolff/minify#Minifier
type noopMinifier struct{}

// Minify copies r into w without transformation.
func (m noopMinifier) Minify(_ *minify.M, w io.Writer, r io.Reader, _ map[string]string) error {
	_, err := io.Copy(w, r)
	return err
}

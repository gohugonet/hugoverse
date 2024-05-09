package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
)

var defaultTdewolffConfig = TdewolffConfig{
	HTML: html.Minifier{
		KeepDocumentTags:    true,
		KeepSpecialComments: true,
		KeepEndTags:         true,
		KeepDefaultAttrVals: true,
		KeepWhitespace:      false,
	},
	CSS: css.Minifier{
		Precision: 0,
		KeepCSS2:  true,
	},
	JS: js.Minifier{
		Version: 2022,
	},
	JSON: json.Minifier{},
	SVG: svg.Minifier{
		KeepComments: false,
		Precision:    0,
	},
	XML: xml.Minifier{
		KeepWhitespace: false,
	},
}

type TdewolffConfig struct {
	HTML html.Minifier
	CSS  css.Minifier
	JS   js.Minifier
	JSON json.Minifier
	SVG  svg.Minifier
	XML  xml.Minifier
}

type MinifyConfig struct {
	// Whether to minify the published output (the HTML written to /public).
	MinifyOutput bool

	DisableHTML bool
	DisableCSS  bool
	DisableJS   bool
	DisableJSON bool
	DisableSVG  bool
	DisableXML  bool

	Tdewolff TdewolffConfig
}

var defaultConfig = MinifyConfig{
	Tdewolff: defaultTdewolffConfig,
}

func DecodeMinifyConfig(p config.Provider) (conf MinifyConfig, err error) {
	m := p.GetStringMap("minify")

	conf = defaultConfig
	// Handle upstream renames.
	if td, found := m["tdewolff"]; found {
		tdm := maps.ToStringMap(td)

		for _, key := range []string{"css", "svg"} {
			if v, found := tdm[key]; found {
				vm := maps.ToStringMap(v)
				ko := "decimal"
				kn := "precision"
				if vv, found := vm[ko]; found {
					if _, found = vm[kn]; !found {
						vvi := cast.ToInt(vv)
						if vvi > 0 {
							vm[kn] = vvi
						}
					}
					delete(vm, ko)
				}
			}
		}

		// keepConditionalComments was renamed to keepSpecialComments
		if v, found := tdm["html"]; found {
			vm := maps.ToStringMap(v)
			ko := "keepconditionalcomments"
			kn := "keepspecialcomments"
			if vv, found := vm[ko]; found {
				// Set keepspecialcomments, if not already set
				if _, found := vm[kn]; !found {
					vm[kn] = cast.ToBool(vv)
				}
				// Remove the old key to prevent deprecation warnings
				delete(vm, ko)
			}
		}

	}

	err = mapstructure.WeakDecode(m, &conf)

	if err != nil {
		return
	}

	return
}

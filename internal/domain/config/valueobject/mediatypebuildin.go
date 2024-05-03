package valueobject

import "github.com/gohugonet/hugoverse/pkg/media"

type BuiltinTypes struct {
	CalendarType   media.Type
	CSSType        media.Type
	SCSSType       media.Type
	SASSType       media.Type
	CSVType        media.Type
	HTMLType       media.Type
	JavascriptType media.Type
	TypeScriptType media.Type
	TSXType        media.Type
	JSXType        media.Type

	JSONType           media.Type
	WebAppManifestType media.Type
	RSSType            media.Type
	XMLType            media.Type
	SVGType            media.Type
	TextType           media.Type
	TOMLType           media.Type
	YAMLType           media.Type

	// Common images types
	PNGType  media.Type
	JPEGType media.Type
	GIFType  media.Type
	TIFFType media.Type
	BMPType  media.Type
	WEBPType media.Type

	// Common font types
	TrueTypeFontType media.Type
	OpenTypeFontType media.Type

	// Common document types
	PDFType      media.Type
	MarkdownType media.Type

	// Common video types
	AVIType  media.Type
	MPEGType media.Type
	MP4Type  media.Type
	OGGType  media.Type
	WEBMType media.Type
	GPPType  media.Type

	// wasm
	WasmType media.Type

	OctetType media.Type
}

var Builtin = BuiltinTypes{
	CalendarType:   media.Type{Type: "text/calendar"},
	CSSType:        media.Type{Type: "text/css"},
	SCSSType:       media.Type{Type: "text/x-scss"},
	SASSType:       media.Type{Type: "text/x-sass"},
	CSVType:        media.Type{Type: "text/csv"},
	HTMLType:       media.Type{Type: "text/html"},
	JavascriptType: media.Type{Type: "text/javascript"},
	TypeScriptType: media.Type{Type: "text/typescript"},
	TSXType:        media.Type{Type: "text/tsx"},
	JSXType:        media.Type{Type: "text/jsx"},

	JSONType:           media.Type{Type: "application/json"},
	WebAppManifestType: media.Type{Type: "application/manifest+json"},
	RSSType:            media.Type{Type: "application/rss+xml"},
	XMLType:            media.Type{Type: "application/xml"},
	SVGType:            media.Type{Type: "images/svg+xml"},
	TextType:           media.Type{Type: "text/plain"},
	TOMLType:           media.Type{Type: "application/toml"},
	YAMLType:           media.Type{Type: "application/yaml"},

	// Common images types
	PNGType:  media.Type{Type: "images/png"},
	JPEGType: media.Type{Type: "images/jpeg"},
	GIFType:  media.Type{Type: "images/gif"},
	TIFFType: media.Type{Type: "images/tiff"},
	BMPType:  media.Type{Type: "images/bmp"},
	WEBPType: media.Type{Type: "images/webp"},

	// Common font types
	TrueTypeFontType: media.Type{Type: "font/ttf"},
	OpenTypeFontType: media.Type{Type: "font/otf"},

	// Common document types
	PDFType:      media.Type{Type: "application/pdf"},
	MarkdownType: media.Type{Type: "text/markdown"},

	// Common video types
	AVIType:  media.Type{Type: "video/x-msvideo"},
	MPEGType: media.Type{Type: "video/mpeg"},
	MP4Type:  media.Type{Type: "video/mp4"},
	OGGType:  media.Type{Type: "video/ogg"},
	WEBMType: media.Type{Type: "video/webm"},
	GPPType:  media.Type{Type: "video/3gpp"},

	// Web assembly.
	WasmType: media.Type{Type: "application/wasm"},

	OctetType: media.Type{Type: "application/octet-stream"},
}

var defaultMediaTypesConfig = map[string]any{
	"text/calendar":   map[string]any{"suffixes": []string{"ics"}},
	"text/css":        map[string]any{"suffixes": []string{"css"}},
	"text/x-scss":     map[string]any{"suffixes": []string{"scss"}},
	"text/x-sass":     map[string]any{"suffixes": []string{"sass"}},
	"text/csv":        map[string]any{"suffixes": []string{"csv"}},
	"text/html":       map[string]any{"suffixes": []string{"html"}},
	"text/javascript": map[string]any{"suffixes": []string{"js", "jsm", "mjs"}},
	"text/typescript": map[string]any{"suffixes": []string{"ts"}},
	"text/tsx":        map[string]any{"suffixes": []string{"tsx"}},
	"text/jsx":        map[string]any{"suffixes": []string{"jsx"}},

	"application/json":          map[string]any{"suffixes": []string{"json"}},
	"application/manifest+json": map[string]any{"suffixes": []string{"webmanifest"}},
	"application/rss+xml":       map[string]any{"suffixes": []string{"xml", "rss"}},
	"application/xml":           map[string]any{"suffixes": []string{"xml"}},
	"images/svg+xml":            map[string]any{"suffixes": []string{"svg"}},
	"text/plain":                map[string]any{"suffixes": []string{"txt"}},
	"application/toml":          map[string]any{"suffixes": []string{"toml"}},
	"application/yaml":          map[string]any{"suffixes": []string{"yaml", "yml"}},

	// Common images types
	"images/png":  map[string]any{"suffixes": []string{"png"}},
	"images/jpeg": map[string]any{"suffixes": []string{"jpg", "jpeg", "jpe", "jif", "jfif"}},
	"images/gif":  map[string]any{"suffixes": []string{"gif"}},
	"images/tiff": map[string]any{"suffixes": []string{"tif", "tiff"}},
	"images/bmp":  map[string]any{"suffixes": []string{"bmp"}},
	"images/webp": map[string]any{"suffixes": []string{"webp"}},

	// Common font types
	"font/ttf": map[string]any{"suffixes": []string{"ttf"}},
	"font/otf": map[string]any{"suffixes": []string{"otf"}},

	// Common document types
	"application/pdf": map[string]any{"suffixes": []string{"pdf"}},
	"text/markdown":   map[string]any{"suffixes": []string{"md", "markdown"}},

	// Common video types
	"video/x-msvideo": map[string]any{"suffixes": []string{"avi"}},
	"video/mpeg":      map[string]any{"suffixes": []string{"mpg", "mpeg"}},
	"video/mp4":       map[string]any{"suffixes": []string{"mp4"}},
	"video/ogg":       map[string]any{"suffixes": []string{"ogv"}},
	"video/webm":      map[string]any{"suffixes": []string{"webm"}},
	"video/3gpp":      map[string]any{"suffixes": []string{"3gpp", "3gp"}},

	// wasm
	"application/wasm": map[string]any{"suffixes": []string{"wasm"}},

	"application/octet-stream": map[string]any{},
}

func init() {
	// Apply delimiter to all.
	for _, m := range defaultMediaTypesConfig {
		m.(map[string]any)["delimiter"] = "."
	}
}

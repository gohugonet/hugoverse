package media

type BuiltinTypes struct {
	CalendarType   Type
	CSSType        Type
	SCSSType       Type
	SASSType       Type
	CSVType        Type
	HTMLType       Type
	JavascriptType Type
	TypeScriptType Type
	TSXType        Type
	JSXType        Type

	JSONType           Type
	WebAppManifestType Type
	RSSType            Type
	XMLType            Type
	SVGType            Type
	TextType           Type
	TOMLType           Type
	YAMLType           Type

	// Common images types
	PNGType  Type
	JPEGType Type
	GIFType  Type
	TIFFType Type
	BMPType  Type
	WEBPType Type

	// Common font types
	TrueTypeFontType Type
	OpenTypeFontType Type

	// Common document types
	PDFType              Type
	MarkdownType         Type
	EmacsOrgModeType     Type
	AsciiDocType         Type
	PandocType           Type
	ReStructuredTextType Type

	// Common video types
	AVIType  Type
	MPEGType Type
	MP4Type  Type
	OGGType  Type
	WEBMType Type
	GPPType  Type

	// wasm
	WasmType Type

	OctetType Type
}

var BuiltinJs = []Type{
	Builtin.JavascriptType,
}

var BuiltinCss = []Type{
	Builtin.CSSType,
	Builtin.SCSSType,
	Builtin.SASSType,
}

var BuiltinImages = []Type{
	Builtin.PNGType,
	Builtin.JPEGType,
	Builtin.GIFType,
	Builtin.TIFFType,
	Builtin.BMPType,
	Builtin.WEBPType,
}

var BuiltinJson = []Type{
	Builtin.JSONType,
}

var Builtin = BuiltinTypes{
	CalendarType:   Type{Type: "text/calendar"},
	CSSType:        Type{Type: "text/css"},
	SCSSType:       Type{Type: "text/x-scss"},
	SASSType:       Type{Type: "text/x-sass"},
	CSVType:        Type{Type: "text/csv"},
	HTMLType:       Type{Type: "text/html"},
	JavascriptType: Type{Type: "text/javascript"},
	TypeScriptType: Type{Type: "text/typescript"},
	TSXType:        Type{Type: "text/tsx"},
	JSXType:        Type{Type: "text/jsx"},

	JSONType:           Type{Type: "application/json"},
	WebAppManifestType: Type{Type: "application/manifest+json"},
	RSSType:            Type{Type: "application/rss+xml"},
	XMLType:            Type{Type: "application/xml"},
	SVGType:            Type{Type: "images/svg+xml"},
	TextType:           Type{Type: "text/plain"},
	TOMLType:           Type{Type: "application/toml"},
	YAMLType:           Type{Type: "application/yaml"},

	// Common images types
	PNGType:  Type{Type: "images/png"},
	JPEGType: Type{Type: "images/jpeg"},
	GIFType:  Type{Type: "images/gif"},
	TIFFType: Type{Type: "images/tiff"},
	BMPType:  Type{Type: "images/bmp"},
	WEBPType: Type{Type: "images/webp"},

	// Common font types
	TrueTypeFontType: Type{Type: "font/ttf"},
	OpenTypeFontType: Type{Type: "font/otf"},

	// Common document types
	PDFType:              Type{Type: "application/pdf"},
	MarkdownType:         Type{Type: "text/markdown"},
	AsciiDocType:         Type{Type: "text/asciidoc"}, // https://github.com/asciidoctor/asciidoctor/issues/2502
	PandocType:           Type{Type: "text/pandoc"},
	ReStructuredTextType: Type{Type: "text/rst"}, // https://docutils.sourceforge.io/FAQ.html#what-s-the-official-mime-type-for-restructuredtext-data
	EmacsOrgModeType:     Type{Type: "text/org"},

	// Common video types
	AVIType:  Type{Type: "video/x-msvideo"},
	MPEGType: Type{Type: "video/mpeg"},
	MP4Type:  Type{Type: "video/mp4"},
	OGGType:  Type{Type: "video/ogg"},
	WEBMType: Type{Type: "video/webm"},
	GPPType:  Type{Type: "video/3gpp"},

	// Web assembly.
	WasmType: Type{Type: "application/wasm"},

	OctetType: Type{Type: "application/octet-stream"},
}

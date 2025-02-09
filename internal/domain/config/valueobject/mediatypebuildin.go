package valueobject

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
	"text/asciidoc":   map[string]any{"suffixes": []string{"adoc", "asciidoc", "ad"}},
	"text/pandoc":     map[string]any{"suffixes": []string{"pandoc", "pdc"}},
	"text/rst":        map[string]any{"suffixes": []string{"rst"}},
	"text/org":        map[string]any{"suffixes": []string{"org"}},

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

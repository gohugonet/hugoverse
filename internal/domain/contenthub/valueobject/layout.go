package valueobject

import (
	"strings"
	"sync"
)

// LayoutDescriptor describes how a layout should be chosen. This is
// typically built from a Page.
type LayoutDescriptor struct {
	Type    string
	Section string

	// E.g. "page", but also used for the _markup render kinds, e.g. "render-images".
	Kind string

	// Comma-separated list of kind variants, e.g. "go,json" as variants which would find "render-codeblock-go.html"
	KindVariants string

	Lang   string
	Layout string
	// LayoutOverride indicates what we should only look for the above layout.
	LayoutOverride bool

	RenderingHook bool
	Baseof        bool

	FormatName string
	Extension  string
}

// LayoutHandler calculates the layout template to use to render a given output type.
type LayoutHandler struct {
	mu    sync.RWMutex
	cache map[layoutCacheKey][]string
}

type layoutCacheKey struct {
	d LayoutDescriptor
	f string
}

// NewLayoutHandler creates a new LayoutHandler.
func NewLayoutHandler() *LayoutHandler {
	return &LayoutHandler{cache: make(map[layoutCacheKey][]string)}
}

// For returns a layout for the given LayoutDescriptor and options.
// Layouts are rendered and cached internally.
func (l *LayoutHandler) For(d LayoutDescriptor) ([]string, error) {
	// We will get lots of requests for the same layouts, so avoid recalculations.
	key := layoutCacheKey{d, d.FormatName}
	l.mu.RLock()
	if cacheVal, found := l.cache[key]; found {
		l.mu.RUnlock()
		return cacheVal, nil
	}
	l.mu.RUnlock()

	layouts := resolvePageTemplate(d)

	layouts = UniqueStringsReuse(layouts)

	l.mu.Lock()
	l.cache[key] = layouts
	l.mu.Unlock()

	return layouts, nil
}

// UniqueStringsReuse returns a slice with any duplicates removed.
// It will modify the input slice.
func UniqueStringsReuse(s []string) []string {
	result := s[:0]
	for i, val := range s {
		var seen bool

		for j := 0; j < i; j++ {
			if s[j] == val {
				seen = true
				break
			}
		}

		if !seen {
			result = append(result, val)
		}
	}
	return result
}

func resolvePageTemplate(d LayoutDescriptor) []string {
	b := &layoutBuilder{d: d}

	if !d.RenderingHook && d.Layout != "" {
		b.addLayoutVariations(d.Layout)
	}
	if d.Type != "" {
		b.addTypeVariations(d.Type)
	}

	switch d.Kind {
	case "page":
		b.addLayoutVariations("single")
		b.addSectionType()
	case "home":
		b.addLayoutVariations("index", "home")
		// Also look in the root
		b.addTypeVariations("")
	case "section":
		if d.Section != "" {
			b.addLayoutVariations(d.Section)
		}
		b.addSectionType()
		b.addKind()
	case "term":
		b.addKind()
		if d.Section != "" {
			b.addLayoutVariations(d.Section)
		}
		b.addLayoutVariations("taxonomy")
		b.addTypeVariations("taxonomy")
		b.addSectionType()
	case "taxonomy":
		if d.Section != "" {
			b.addLayoutVariations(d.Section + ".terms")
		}
		b.addSectionType()
		b.addLayoutVariations("terms")
		// For legacy reasons this is deliberately put last.
		b.addKind()
	case "404":
		b.addLayoutVariations("404")
		b.addTypeVariations("")
	}

	if d.Baseof || d.Kind != "404" {
		// Most have _default in their lookup path
		b.addTypeVariations("_default")
	}

	if d.isList() {
		// Add the common list type
		b.addLayoutVariations("list")
	}

	if d.Baseof {
		b.addLayoutVariations("baseof")
	}

	layouts := b.resolveVariations()

	return layouts
}

type layoutBuilder struct {
	layoutVariations []string
	typeVariations   []string
	d                LayoutDescriptor
}

func (l *layoutBuilder) addLayoutVariations(vars ...string) {
	for _, layoutVar := range vars {
		if l.d.Baseof && layoutVar != "baseof" {
			l.layoutVariations = append(l.layoutVariations, layoutVar+"-baseof")
			continue
		}
		if !l.d.RenderingHook && !l.d.Baseof && l.d.LayoutOverride && layoutVar != l.d.Layout {
			continue
		}
		l.layoutVariations = append(l.layoutVariations, layoutVar)
	}
}

func (l *layoutBuilder) addTypeVariations(vars ...string) {
	for _, typeVar := range vars {
		if !reservedSections[typeVar] {
			if l.d.RenderingHook {
				typeVar = typeVar + renderingHookRoot
			}
			l.typeVariations = append(l.typeVariations, typeVar)
		}
	}
}

// These may be used as content sections with potential conflicts. Avoid that.
var reservedSections = map[string]bool{
	"shortcodes": true,
	"partials":   true,
}

const renderingHookRoot = "/_markup"

func (l *layoutBuilder) addSectionType() {
	if l.d.Section != "" {
		l.addTypeVariations(l.d.Section)
	}
}

func (l *layoutBuilder) addKind() {
	l.addLayoutVariations(l.d.Kind)
	l.addTypeVariations(l.d.Kind)
}

func (d LayoutDescriptor) isList() bool {
	return !d.RenderingHook && d.Kind != "page" && d.Kind != "404"
}

func (l *layoutBuilder) resolveVariations() []string {
	var layouts []string

	var variations []string
	name := strings.ToLower(l.d.FormatName)
	variations = append(variations, name)

	variations = append(variations, "")

	for _, typeVar := range l.typeVariations {
		for _, variation := range variations {
			for _, layoutVar := range l.layoutVariations {
				if variation == "" && layoutVar == "" {
					continue
				}

				s := constructLayoutPath(typeVar, layoutVar, variation, l.d.Extension)
				if s != "" {
					layouts = append(layouts, s)
				}
			}
		}
	}

	return layouts
}

// constructLayoutPath constructs a layout path given a type, layout,
// variations, and extension.  The path constructed follows the pattern of
// type/layout.variations.extension.  If any value is empty, it will be left out
// of the path construction.
//
// RelPath construction requires at least 2 of 3 out of layout, variations, and extension.
// If more than one of those is empty, an empty string is returned.
func constructLayoutPath(typ, layout, variations, extension string) string {
	// we already know that layout and variations are not both empty because of
	// checks in resolveVariants().
	if extension == "" && (layout == "" || variations == "") {
		return ""
	}

	// Commence valid path construction...

	var (
		p       strings.Builder
		needDot bool
	)

	if typ != "" {
		p.WriteString(typ)
		p.WriteString("/")
	}

	if layout != "" {
		p.WriteString(layout)
		needDot = true
	}

	if variations != "" {
		if needDot {
			p.WriteString(".")
		}
		p.WriteString(variations)
		needDot = true
	}

	if extension != "" {
		if needDot {
			p.WriteString(".")
		}
		p.WriteString(extension)
	}

	return p.String()
}

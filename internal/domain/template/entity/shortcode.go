package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	"strings"
)

const (
	textTmplNamePrefix = "_text/"

	shortcodesPathPrefix = "shortcodes/"
	internalPathPrefix   = "_internal/"
	embeddedPathPrefix   = "_embedded/"
	baseFileBase         = "baseof"

	numTemplateVariants = 3
)

type Shortcode struct {
	// shortcodes maps shortcode name to template variants
	// (language, output format etc.) of that shortcode.
	shortcodes map[string]*shortcodeTemplates
}

func (t *Shortcode) addShortcodeVariant(ts *valueobject.State) {
	name := ts.Name()
	base := templateBaseName(template.TypeShortcode, name)

	shortcodename, variants := templateNameAndVariants(base)

	templs, found := t.shortcodes[shortcodename]
	if !found {
		templs = &shortcodeTemplates{}
		t.shortcodes[shortcodename] = templs
	}

	sv := shortcodeVariant{variants: variants, ts: ts}

	i := templs.indexOf(variants)

	if i != -1 {
		// Only replace if it's an override of an internal template.
		if !isInternal(name) {
			templs.variants[i] = sv
		}
	} else {
		templs.variants = append(templs.variants, sv)
	}
}

// resolves _internal/shortcodes/param.html => param.html etc.
func templateBaseName(typ template.Type, name string) string {
	name = strings.TrimPrefix(name, internalPathPrefix)
	switch typ {
	case template.TypeShortcode:
		return strings.TrimPrefix(name, shortcodesPathPrefix)
	default:
		panic("not implemented")
	}
}

func templateNameAndVariants(name string) (string, []string) {
	variants := make([]string, numTemplateVariants)

	parts := strings.Split(name, ".")

	if len(parts) <= 1 {
		// No variants.
		return name, variants
	}

	name = parts[0]
	parts = parts[1:]
	lp := len(parts)
	start := len(variants) - lp

	for i, j := start, 0; i < len(variants); i, j = i+1, j+1 {
		variants[i] = parts[j]
	}

	if lp > 1 && lp < len(variants) {
		for i := lp - 1; i > 0; i-- {
			variants[i-1] = variants[i]
		}
	}

	if lp == 1 {
		// Suffix only. Duplicate it into the output format field to
		// make HTML win over AMP.
		variants[len(variants)-2] = variants[len(variants)-1]
	}

	return name, variants
}

func isInternal(name string) bool {
	return strings.HasPrefix(name, internalPathPrefix)
}

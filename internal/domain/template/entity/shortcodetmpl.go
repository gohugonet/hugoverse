package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	"strings"
)

type shortcodeVariant struct {
	// The possible variants: lang, outFormat, suffix
	// gtag
	// gtag.html
	// gtag.no.html
	// gtag.no.amp.html
	// A slice of length NumTemplateVariants.
	variants []string

	ts *valueobject.State
}

type shortcodeTemplates struct {
	variants []shortcodeVariant
}

func (s *shortcodeTemplates) indexOf(variants []string) int {
L:
	for i, v1 := range s.variants {
		for i, v2 := range v1.variants {
			if v2 != variants[i] {
				continue L
			}
		}
		return i
	}
	return -1
}

func (s *shortcodeTemplates) fromVariants(variants TemplateVariants) (shortcodeVariant, bool) {
	return s.fromVariantsSlice([]string{
		variants.Language,
		strings.ToLower(variants.OutputFormat.Name),
		variants.OutputFormat.MediaType.FirstSuffix.Suffix,
	})
}

func (s *shortcodeTemplates) fromVariantsSlice(variants []string) (shortcodeVariant, bool) {
	var (
		bestMatch       shortcodeVariant
		bestMatchWeight int
	)

	for _, variant := range s.variants {
		w := s.compareVariants(variants, variant.variants)
		if bestMatchWeight == 0 || w > bestMatchWeight {
			bestMatch = variant
			bestMatchWeight = w
		}
	}

	return bestMatch, true
}

// calculate a weight for two string slices of same length.
// higher value means "better match".
func (s *shortcodeTemplates) compareVariants(a, b []string) int {
	weight := 0
	k := len(a)
	for i, av := range a {
		bv := b[i]
		if av == bv {
			// Add more weight to the left side (language...).
			weight = weight + k - i
		} else {
			weight--
		}
	}
	return weight
}

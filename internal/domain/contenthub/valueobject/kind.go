package valueobject

import (
	"sort"
	"strings"
)

const (
	KindPage = "page"

	// The rest are node types; home page, sections etc.

	KindHome    = "home"
	KindSection = "section"

	// Note that before Hugo 0.73 these were confusingly named
	// taxonomy (now: term)
	// taxonomyTerm (now: taxonomy)
	KindTaxonomy = "taxonomy"
	KindTerm     = "term"

	// The following are (currently) temporary nodes,
	// i.e. nodes we create just to render in isolation.
	KindRSS          = "rss"
	KindSitemap      = "sitemap"
	KindSitemapIndex = "sitemapindex"
	KindRobotsTXT    = "robotstxt"
	KindStatus404    = "404"
)

var (
	// This is all the kinds we can expect to find in .Site.Pages.
	AllKindsInPages []string
	// This is all the kinds, including the temporary ones.
	AllKinds []string
)

func init() {
	for k := range kindMapMain {
		AllKindsInPages = append(AllKindsInPages, k)
		AllKinds = append(AllKinds, k)
	}

	for k := range kindMapTemporary {
		AllKinds = append(AllKinds, k)
	}

	// Sort the slices for determinism.
	sort.Strings(AllKindsInPages)
	sort.Strings(AllKinds)
}

var kindMapMain = map[string]string{
	KindPage:     KindPage,
	KindHome:     KindHome,
	KindSection:  KindSection,
	KindTaxonomy: KindTaxonomy,
	KindTerm:     KindTerm,

	// Legacy, pre v0.53.0.
	"taxonomyterm": KindTaxonomy,
}

var kindMapTemporary = map[string]string{
	KindRSS:       KindRSS,
	KindSitemap:   KindSitemap,
	KindRobotsTXT: KindRobotsTXT,
	KindStatus404: KindStatus404,
}

// GetKindMain gets the page kind given a string, empty if not found.
// Note that this will not return any temporary kinds (e.g. robotstxt).
func GetKindMain(s string) string {
	return kindMapMain[strings.ToLower(s)]
}

// GetKindAny gets the page kind given a string, empty if not found.
func GetKindAny(s string) string {
	if pkind := GetKindMain(s); pkind != "" {
		return pkind
	}
	return kindMapTemporary[strings.ToLower(s)]
}

// IsBranch returns whether the given kind is a branch node.
func IsBranch(kind string) bool {
	switch kind {
	case KindHome, KindSection, KindTaxonomy, KindTerm:
		return true
	default:
		return false
	}
}

// IsDeprecatedAndReplacedWith returns the new kind if the given kind is deprecated.
func IsDeprecatedAndReplacedWith(s string) string {
	s = strings.ToLower(s)

	switch s {
	case "taxonomyterm":
		return KindTaxonomy
	default:
		return ""
	}
}

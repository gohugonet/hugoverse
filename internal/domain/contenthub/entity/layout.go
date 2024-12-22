package entity

import "fmt"

const (
	LayoutSection = "section.html"
	LayoutList    = "list.html"
	LayoutIndex   = "index.html"

	InternalFolder = "_internal"

	DefaultFolder   = "_default"
	DefaultIndex    = "_default/index.html"
	DefaultList     = DefaultFolder + "/" + LayoutList
	DefaultPage     = "_default/single.html"
	DefaultSection  = DefaultFolder + "/" + LayoutSection
	DefaultTaxonomy = "_default/taxonomy.html"
	DefaultTerm     = "_default/term.html"
	DefaultBaseof   = "_default/baseof.html"

	TaxonomyTaxonomy = "taxonomy/taxonomy.html"
	TaxonomyList     = "taxonomy" + "/" + LayoutList

	TermTerm = "term/term.html"
	TermTag  = "taxonomy/tag.html"
	TermList = "taxonomy" + "/" + LayoutList

	Sitemap                = "sitemap.xml"
	DefaultSitemap         = DefaultFolder + "/" + "sitemap.xml"
	InternalDefaultSitemap = InternalFolder + "/" + DefaultFolder + "/" + "sitemap.xml"
)

type Layout struct{}

func (l *Layout) home() []string {
	return []string{
		LayoutIndex,
		DefaultIndex,
		DefaultList,
	}
}

func (l *Layout) section(section string) []string {
	return []string{
		fmt.Sprintf("%s/%s", section, LayoutSection),
		fmt.Sprintf("%s/%s", section, LayoutList),
		DefaultSection,
		DefaultList,
		DefaultIndex,
	}
}

func (l *Layout) page() []string {
	return []string{
		DefaultPage,
		DefaultBaseof,
	}
}

func (l *Layout) taxonomy() []string {
	return []string{
		TaxonomyTaxonomy,
		TaxonomyList,
		DefaultList,
		DefaultTaxonomy,
	}
}

func (l *Layout) term() []string {
	return []string{
		TermTag,
		TermTerm,
		TermList,
		DefaultList,
		DefaultTerm,
	}
}

func (l *Layout) standalone404() []string {
	return []string{
		"404.html",
	}
}

func (l *Layout) standaloneSitemap() []string {
	return []string{
		Sitemap,
		DefaultSitemap,
		InternalDefaultSitemap,
	}
}

package entity

import "fmt"

const (
	LayoutSection = "section.html"
	LayoutList    = "list.html"
	LayoutIndex   = "index.html"

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
	TermList = "taxonomy" + "/" + LayoutList
)

type Layout struct {
}

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
		DefaultTaxonomy,
		DefaultList,
	}
}

func (l *Layout) term() []string {
	return []string{
		TermTerm,
		TermList,
		DefaultTerm,
		DefaultList,
	}
}

func (l *Layout) standalone(name string) []string {
	return []string{
		fmt.Sprintf("%s.html", name),
	}
}

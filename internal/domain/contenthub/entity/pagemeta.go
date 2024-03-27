package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"sync"
)

type pageMeta struct {
	// kind is the discriminator that identifies the different page types
	// in the different page collections. This can, as an example, be used
	// to to filter regular pages, find sections etc.
	// Kind will, for the pages available to the templates, be one of:
	// page, home, section, taxonomy and term.
	// It is of string type to make it easy to reason about in
	// the templates.
	kind string

	// Params contains configuration defined in the params section of page frontmatter.
	params map[string]any

	title     string
	linkTitle string

	summary string

	resourcePath string

	weight int

	markup      string
	contentType string

	// whether the content is in a CJK language.
	isCJKLanguage bool

	layout string

	aliases []string

	description string
	keywords    []string

	//urlPaths pagemeta.URLPath

	// Set if this page is bundled inside another.
	bundled bool

	// A key that maps to translation(s) of this page. This value is fetched
	// from the page front matter.
	translationKey string

	// This is the raw front matter metadata that is going to be assigned to
	// the Resources above.
	resourcesMetadata []map[string]any

	sections []string

	contentConverterInit sync.Once
	contentConverter     contenthub.Converter

	f contenthub.File
}

func (p *pageMeta) applyDefaultValues() { // buildConfig, markup, title
	if p.markup == "" {
		p.markup = "markdown"
	}

	p.title = "hardcode title"
}

func (p *pageMeta) File() contenthub.File {
	return p.f
}

func (p *pageMeta) Kind() string {
	return p.kind
}

func (p *pageMeta) SectionsEntries() []string {
	return p.sections
}

const defaultContentType = "page"

func (p *pageMeta) Type() string {
	return defaultContentType
}

func (p *pageMeta) Layout() string {
	return p.layout
}

// The output formats this page will be rendered to.
func (p *pageMeta) outputFormats() valueobject.Formats {
	// TODO
	return nil
	//return p.s.OutputFormats[p.Kind()]
}

func (p *pageMeta) noLink() bool {
	return false
}

func (p *pageMeta) newContentConverter(ps *pageState, markup string) (contenthub.Converter, error) {
	//TODO

	return nil, nil
}

package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"path"
	"strings"
)

// PageCollections contains the page collections for a site.
type PageCollections struct {
	PageMap *PageMap
}

type PageMap struct {
	*ContentMap
}

func (m *PageMap) Assemble() error {
	if err := m.CreateMissingNodes(); err != nil {
		return err
	}

	if err := m.AssemblePages(); err != nil {
		return err
	}

	// Handle any new sections created in the step above.
	if err := m.AssembleSections(); err != nil {
		return err
	}
	return nil
}

func (m *PageMap) AssemblePages() error {
	var err error
	if err = m.AssembleSections(); err != nil {
		return err
	}

	m.Pages.Walk(func(k string, v any) bool {
		n := v.(*contentNode)
		if n.p != nil {
			return false
		}

		_, parent := m.getSection(k)
		if parent == nil {
			panic(fmt.Sprintf("BUG: parent not set for %q", k))
		}

		n.p, err = m.newPageFromContentNode(n)
		fmt.Printf(">>> AssemblePages 000 %+v\n", n.p)

		if err != nil {
			return true
		}

		return false
	})

	return err
}

func (m *PageMap) AssembleSections() error {
	m.Sections.Walk(func(k string, v any) bool {
		n := v.(*contentNode)
		sections := m.splitKey(k)

		kind := contenthub.KindSection
		if k == "/" {
			kind = contenthub.KindHome
		}
		if n.fi != nil {
			panic("assembleSections newPageFromContentNode not ready")
		} else {
			n.p = newPage(n, kind, sections...)
			fmt.Println(">>> assembleSections new page 999", sections, k, n.p)
		}

		return false
	})

	return nil
}

func (m *PageMap) newPageFromContentNode(n *contentNode) (*pageState, error) {
	if n.fi == nil {
		panic("FileInfo must (currently) be set")
	}

	f, err := newFileInfo(n.fi)
	if err != nil {
		return nil, err
	}

	meta := n.fi.Meta()
	content := func() (io.ReadSeekCloser, error) {
		return meta.Open()
	}

	sections := m.sectionsFromFile(f)
	kind := m.kindFromFileInfoOrSections(f, sections)

	// TODO, site in meta provider
	metaProvider := &pageMeta{kind: kind, sections: sections, bundled: false, f: f}
	ps, err := newPageBase(metaProvider)
	if err != nil {
		return nil, err
	}

	n.p = ps
	r, err := content()
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// .md parseResult
	// TODO: parser works way
	parseResult, err := pageparser.Parse(
		r,
		pageparser.Config{EnableEmoji: false},
	)
	if err != nil {
		return nil, err
	}

	ps.pageContent = pageContent{
		source: rawPageContent{
			parsed:         parseResult,
			posMainContent: -1,
			posSummaryEnd:  -1,
			posBodyStart:   -1,
		},
	}

	if err := ps.mapContent(metaProvider); err != nil {
		return nil, err
	}
	metaProvider.applyDefaultValues()
	ps.init.Add(func() (any, error) {
		fmt.Printf("TODO init page start : %+v\n", ps)

		makeOut := func() *pageOutput {
			return newPageOutput()
		}

		// Prepare output formats for all sites.
		// We do this even if this page does not get rendered on
		// its own. It may be referenced via .Site.GetPage and
		// it will then need an output format.
		ps.pageOutputs = make([]*pageOutput, 1)
		po := makeOut()
		ps.pageOutputs[0] = po

		contentProvider, err := newPageContentOutput(ps)
		if err != nil {
			return nil, err
		}
		po.initContentProvider(contentProvider)

		return nil, nil
	})

	return ps, nil
}

func (m *PageMap) sectionsFromFile(fi contenthub.File) []string {
	dirname := fi.Dir()

	dirname = strings.Trim(dirname, paths.FilePathSeparator)
	if dirname == "" {
		return nil
	}
	parts := strings.Split(dirname, paths.FilePathSeparator)

	if fii, ok := fi.(*fileInfo); ok {
		if len(parts) > 0 && fii.FileInfo().Meta().Classifier == fsVO.ContentClassLeaf {
			// my-section/mybundle/index.md => my-section
			return parts[:len(parts)-1]
		}
	}

	return parts
}

func (m *PageMap) kindFromFileInfoOrSections(fi *fileInfo, sections []string) string {
	if fi.TranslationBaseName() == "_index" {
		if fi.Dir() == "" {
			return contenthub.KindHome
		}
		return m.kindFromSections(sections)
	}

	return contenthub.KindPage
}

func (m *PageMap) kindFromSections(sections []string) string {
	if len(sections) == 0 {
		return contenthub.KindHome
	}

	return m.kindFromSectionPath(path.Join(sections...))
}

func (m *PageMap) kindFromSectionPath(sectionPath string) string {
	return contenthub.KindSection
}

func (m *PageMap) splitKey(k string) []string {
	if k == "" || k == "/" {
		return nil
	}

	return strings.Split(k, "/")[1:]
}

// withEveryBundlePage applies fn to every Page, including those bundled inside
// leaf bundles.
func (m *PageMap) withEveryBundlePage(fn func(p *pageState) bool) {
	m.BundleTrees.Walk(func(s string, n *contentNode) bool {
		if n.p != nil {
			return fn(n.p)
		}
		return false
	})
}

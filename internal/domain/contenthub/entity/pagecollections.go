package entity

import (
	"fmt"
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/loggers"
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
	*ContentSpec

	Log loggers.Logger
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
			n.p = newPage(m.ContentSpec, n, kind, sections...)
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

	f, err := valueobject.NewFileInfo(n.fi)
	if err != nil {
		return nil, err
	}

	meta := n.fi.Meta()
	content := func() (io.ReadSeekCloser, error) {
		return meta.Open()
	}

	sections := m.sectionsFromFile(f)
	kind := m.kindFromFileInfoOrSections(f, sections)

	metaProvider := &pageMeta{
		kind: kind, sections: sections, bundled: false, f: f,
		contentCovertProvider: m.ContentSpec,
	}

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

		po := newPageOutput()
		ps.pageOutput = po

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

	if fii, ok := fi.(*valueobject.File); ok {
		if len(parts) > 0 && fii.FileInfo().Meta().Classifier == fsVO.ContentClassLeaf {
			// my-section/mybundle/index.md => my-section
			return parts[:len(parts)-1]
		}
	}

	return parts
}

func (m *PageMap) kindFromFileInfoOrSections(fi *valueobject.File, sections []string) string {
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

func (m *PageMap) AddFi(fi fs.FileMetaInfo) error {
	if fi.IsDir() {
		return nil
	}

	insertResource := func(fim fs.FileMetaInfo) error {
		pi := fi.Path()
		key := pi.Base()
		tree := m.treeResources

		commit := tree.Lock(true)
		defer commit()

		r := func() (hugio.ReadSeekCloser, error) {
			return fim.Meta().Open()
		}

		var rs *resourceSource
		if pi.IsContent() {
			// Create the page now as we need it at assemembly time.
			// The other resources are created if needed.
			pageResource, pi, err := m.s.h.newPage(
				&pageMeta{
					f:        source.NewFileInfo(fim),
					pathInfo: pi,
					bundled:  true,
				},
			)
			if err != nil {
				return err
			}
			if pageResource == nil {
				// Disabled page.
				return nil
			}
			key = pi.Base()

			rs = &resourceSource{r: pageResource}
		} else {
			rs = &resourceSource{path: pi, opener: r, fi: fim}
		}

		tree.InsertIntoValuesDimension(key, rs)

		return nil
	}

	pi := fi.Path()

	switch pi.BundleType() {
	case paths.PathTypeFile, paths.PathTypeContentResource:
		m.s.Log.Trace(logg.StringFunc(
			func() string {
				return fmt.Sprintf("insert resource: %q", fi.Meta().Filename)
			},
		))
		if err := insertResource(fi); err != nil {
			return err
		}
	default:
		m.s.Log.Trace(logg.StringFunc(
			func() string {
				return fmt.Sprintf("insert bundle: %q", fi.Meta().Filename)
			},
		))
		// A content file.
		p, pi, err := m.s.h.newPage(
			&pageMeta{
				f:        source.NewFileInfo(fi),
				pathInfo: pi,
				bundled:  false,
			},
		)
		if err != nil {
			return err
		}
		if p == nil {
			// Disabled page.
			return nil
		}

		m.treePages.InsertWithLock(pi.Base(), p)

	}
	return nil
}

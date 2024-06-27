package entity

import (
	"fmt"
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/paths"
)

// PageCollections contains the page collections for a site.
type PageCollections struct {
	PageMap *PageMap
}

type PageMap struct {
	*ContentSpec

	// Main storage for all pages.
	*PageTrees

	Cache *valueobject.Cache

	LangSvc     contenthub.LangService
	TaxonomySvc contenthub.TaxonomyService
	TemplateSvc contenthub.Template

	Log loggers.Logger
}

func (m *PageMap) AddFi(fi fs.FileMetaInfo) error {
	if fi.IsDir() {
		return nil
	}

	pi := paths.Parse(fi.Component(), fi.FileName())

	insertResource := func(fim fs.FileMetaInfo) error {
		key := pi.Base()
		tree := m.TreeResources

		commit := tree.Lock(true)
		defer commit()

		r := func() (io.ReadSeekCloser, error) {
			return fim.Open()
		}

		ps, err := newPageSource(fim, m.Cache)
		if err != nil {
			return err
		}

		if pi.IsContent() {
			// Create the page now as we need it at assemembly time.
			// The other resources are created if needed.
			p, err := newBundledPage(ps, m.LangSvc, m.TaxonomySvc, m.TemplateSvc)
			if err != nil {
				return err
			}

			if err != nil {
				return err
			}
			if p == nil {
				// Disabled page.
				return nil
			}
			key = pi.Base()

			tree.InsertIntoValuesDimension(key, p)
		} else {
			tree.InsertIntoValuesDimension(key, ps)
		}

		return nil
	}

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
		p, pi, err := newPage(
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

		m.TreePages.InsertWithLock(pi.Base(), p)

	}
	return nil
}

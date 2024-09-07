package entity

import (
	"context"
	"fmt"
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"path"
)

type PageMap struct {
	// Main storage for all pages.
	*PageTrees

	Cache *Cache

	PageBuilder *PageBuilder

	Log loggers.Logger
}

func (m *PageMap) PageHome() contenthub.Page {
	return m.PageBuilder.Section.home
}

func (m *PageMap) InsertResourceNode(key string, node *PageTreesNode) {
	tree := m.TreeResources

	commit := tree.Lock(true)
	defer commit()

	tree.InsertIntoValuesDimension(key, node)
}

func (m *PageMap) AddFi(f *valueobject.File) error {
	if f.IsDir() {
		return nil
	}

	ps, err := newPageSource(f, m.Cache)
	if err != nil {
		return err
	}

	key := ps.Path().Base()

	switch ps.BundleType {
	case valueobject.BundleTypeFile:
		m.Log.Trace(logg.StringFunc(
			func() string {
				return fmt.Sprintf("insert resource file: %q", f.FileName())
			},
		))
		m.InsertResourceNode(key, newPageTreesNode(ps))

	case valueobject.BundleTypeContentResource:
		m.Log.Trace(logg.StringFunc(
			func() string {
				return fmt.Sprintf("insert content resource: %q", f.FileName())
			},
		))
		p, err := m.PageBuilder.WithSource(ps).Build()
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

		m.InsertResourceNode(key, newPageTreesNode(p))

	default:
		m.Log.Trace(logg.StringFunc(
			func() string {
				return fmt.Sprintf("insert bundle: %q", f.FileName())
			},
		))
		// A content file.
		p, err := m.PageBuilder.WithSource(ps).Build()
		if err != nil {
			return err
		}

		m.TreePages.InsertWithLock(ps.Path().Base(), newPageTreesNode(p))

	}
	return nil
}

func (m *PageMap) Assemble() error {
	if err := m.assembleStructurePages(); err != nil {
		return err
	}

	if err := m.applyAggregates(); err != nil {
		return err
	}

	if err := m.cleanPages(); err != nil {
		return err
	}

	if err := m.assembleTerms(); err != nil {
		return err
	}

	return nil
}

func (m *PageMap) assembleTerms() error {
	if err := m.PageBuilder.Term.Assemble(m.TreePages, m.TreeTaxonomyEntries, m.PageBuilder); err != nil {
		return err
	}

	return nil
}

func (m *PageMap) cleanPages() error {
	// TODO: clean all the draft, expired, scheduled in the future pages
	return nil
}

func (m *PageMap) applyAggregates() error {

	// TODO
	// Apply cascade Aggregates to pages
	// Apply Dates to home, section, and meta changed pages
	// Apply cascade to source page
	// Use linked list to connect all the cascade in the same path

	// Restore all the changed pages in the cache

	return nil
}

func (m *PageMap) assembleStructurePages() error {

	if err := m.addMissingTaxonomies(); err != nil {
		return err
	}

	if err := m.PageBuilder.Section.Assemble(m.TreePages, m.PageBuilder); err != nil {
		return err
	}

	if err := m.addMissingStandalone(); err != nil {
		return err
	}

	return nil
}

func (m *PageMap) addMissingTaxonomies() error {
	tree := m.TreePages

	commit := tree.Lock(true)
	defer commit()

	if err := m.PageBuilder.Taxonomy.Assemble(tree, m.PageBuilder); err != nil {
		return err
	}

	return nil
}

func (m *PageMap) addMissingStandalone() error {
	tree := m.TreePages

	commit := tree.Lock(true)
	defer commit()

	if err := m.PageBuilder.Standalone.Assemble(tree, m.PageBuilder); err != nil {
		return err
	}

	return nil
}

func (m *PageMap) getResourcesForPage(ps contenthub.Page) ([]contenthub.PageSource, error) {
	var res []contenthub.PageSource

	if err := m.forEachResourceInPage(ps, doctree.LockTypeNone, false,
		func(resourceKey string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {
			rs, found := n.getResource()
			if found {
				res = append(res, rs)
			}
			return false, nil
		}); err != nil {

		return nil, err
	}

	return res, nil
}

func (m *PageMap) forEachResourceInPage(ps contenthub.Page, lockType doctree.LockType, exact bool,
	handle func(resourceKey string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error)) error {

	keyPage := ps.Path().Path()
	if keyPage == "/" {
		keyPage = ""
	}
	prefix := paths.AddTrailingSlash(keyPage)
	isBranch := ps.Kind() != valueobject.KindPage

	rw := &doctree.NodeShiftTreeWalker[*PageTreesNode]{
		Tree:     m.TreeResources.Shape(0, ps.PageIdentity().PageLanguageIndex()),
		Prefix:   prefix,
		LockType: lockType,
		Exact:    exact,
	}

	rw.Handle = func(resourceKey string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {
		if isBranch {
			ownerKey, _ := m.TreePages.LongestPrefixAll(resourceKey)
			if ownerKey != keyPage && path.Dir(ownerKey) != path.Dir(resourceKey) {
				// Stop walking downwards, someone else owns this resource.
				rw.SkipPrefix(ownerKey + "/")
				return false, nil
			}
		}
		return handle(resourceKey, n, match)
	}

	return rw.Walk(context.Background())
}

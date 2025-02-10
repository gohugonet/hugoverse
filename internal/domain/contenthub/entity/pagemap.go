package entity

import (
	"context"
	"fmt"
	"github.com/bep/logg"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/pkg/cache/dynacache"
	"github.com/mdfriday/hugoverse/pkg/doctree"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/mdfriday/hugoverse/pkg/paths"
	"path"
	"strings"
)

type PageMap struct {
	// Main storage for all pages.
	*PageTrees

	// Used for simple page lookups by name, e.g. "mypage.md" or "mypage".
	pageReverseIndex *contentTreeReverseIndex

	Cache *Cache

	PageBuilder *PageBuilder

	Log loggers.Logger
}

func (m *PageMap) SetupReverseIndex() {
	m.pageReverseIndex = &contentTreeReverseIndex{
		initFn: func(rm map[any]*PageTreesNode) {
			add := func(k string, n *PageTreesNode) {
				existing, found := rm[k]
				if found && existing != ambiguousContentNode {
					rm[k] = ambiguousContentNode
				} else if !found {
					rm[k] = n
				}
			}

			w := &doctree.NodeShiftTreeWalker[*PageTreesNode]{
				Tree:     m.TreePages,
				LockType: doctree.LockTypeRead,
				Handle: func(s string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {
					if n != nil {
						p, found := n.getPage()
						if !found {
							return false, nil
						}
						if p.PageFile() != nil {
							add(p.Paths().BaseNameNoIdentifier(), n)
						}
					}

					return false, nil
				},
			}

			if err := w.Walk(context.Background()); err != nil {
				panic(err)
			}
		},
		contentTreeReverseIndexMap: &contentTreeReverseIndexMap{},
	}
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

	key := ps.Paths().Base()

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

		m.TreePages.InsertWithLock(ps.Paths().Base(), newPageTreesNode(p))

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

	for _, idx := range m.PageBuilder.LangSvc.LanguageIndexes() {
		tree := m.TreePages.Shape(0, idx)
		if err := m.PageBuilder.Section.Assemble(tree, m.PageBuilder, idx); err != nil {
			return err
		}
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

	keyPage := ps.Paths().Base()
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

func (m *PageMap) getPagesInSection(langIndex int, q pageMapQueryPagesInSection) contenthub.Pages {
	cacheKey := q.Key()
	tree := m.TreePages.Shape(0, langIndex)

	pages, err := m.getOrCreatePagesFromCache(nil, cacheKey, func(string) (contenthub.Pages, error) {
		prefix := paths.AddTrailingSlash(q.Path)

		var (
			pas         contenthub.Pages
			otherBranch string
		)

		include := q.Include
		if include == nil {
			include = pagePredicates.ShouldListLocal
		}

		w := &doctree.NodeShiftTreeWalker[*PageTreesNode]{
			Tree:   tree,
			Prefix: prefix,
		}

		w.Handle = func(key string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {
			if q.Recursive {
				p, found := n.getPage()
				if found && include(p) {
					pas = append(pas, p)
				}

				return false, nil
			}

			p, found := n.getPage()
			if found && include(p) {
				pas = append(pas, p)
			}

			if !p.IsPage() {
				currentBranch := key + "/"
				if otherBranch == "" || otherBranch != currentBranch {
					w.SkipPrefix(currentBranch)
				}
				otherBranch = currentBranch
			}
			return false, nil
		}

		err := w.Walk(context.Background())

		if err == nil {
			if q.IncludeSelf {
				if n := tree.Get(q.Path); n != nil {
					p, found := n.getPage()
					if found && include(p) {
						pas = append(pas, p)
					}
				}
			}
			valueobject.SortByWeight(pas)
		}

		return pas, err
	})
	if err != nil {
		panic(err)
	}

	return pages
}

func (m *PageMap) getOrCreatePagesFromCache(
	cache *dynacache.Partition[string, contenthub.Pages],
	key string, create func(string) (contenthub.Pages, error),
) (contenthub.Pages, error) {

	if cache == nil {
		cache = m.Cache.CachePages1
	}
	return cache.GetOrCreate(key, create)
}

func (m *PageMap) getPagesWithTerm(q pageMapQueryPagesBelowPath) contenthub.Pages {
	key := q.Key()

	v, err := m.Cache.CachePages1.GetOrCreate(key, func(string) (contenthub.Pages, error) {
		var pas contenthub.Pages
		include := q.Include
		if include == nil {
			include = pagePredicates.ShouldListLocal
		}

		err := m.TreeTaxonomyEntries.WalkPrefix(
			doctree.LockTypeNone,
			paths.AddTrailingSlash(q.Path),
			func(s string, n *WeightedTermTreeNode) (bool, error) {
				p, found := n.getPage()
				if found && include(p) {
					pas = append(pas, p)
				}

				return false, nil
			},
		)
		if err != nil {
			m.Log.Errorf("getPagesWithTerm error: %v", err)
			return nil, err
		}

		valueobject.SortByDefault(pas)

		return pas, nil
	})
	if err != nil {
		m.Log.Errorf("getPagesWithTerm: %v", err)
		panic(err)
	}

	return v
}

func (m *PageMap) getTermsForPageInTaxonomy(base, taxonomy string) contenthub.Pages {
	prefix := paths.AddLeadingSlash(taxonomy)

	v, err := m.Cache.CachePages1.GetOrCreate(prefix+base, func(string) (contenthub.Pages, error) {
		var pas contenthub.Pages

		err := m.TreeTaxonomyEntries.WalkPrefix(
			doctree.LockTypeNone,
			paths.AddTrailingSlash(prefix),
			func(s string, n *WeightedTermTreeNode) (bool, error) {
				if strings.HasSuffix(s, base) {
					pas = append(pas, n.term)
				}
				return false, nil
			},
		)
		if err != nil {
			return nil, err
		}

		valueobject.SortByDefault(pas)

		return pas, nil
	})
	if err != nil {
		panic(err)
	}

	return v
}

func (m *PageMap) getSections(langIndex int, prefix string) contenthub.Pages {
	var (
		pages               contenthub.Pages
		currentBranchPrefix string
		tree                = m.TreePages.Shape(0, langIndex)
	)

	w := &doctree.NodeShiftTreeWalker[*PageTreesNode]{
		Tree:   tree,
		Prefix: prefix,
	}
	w.Handle = func(ss string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {
		p, found := n.getPage()
		if !found {
			return false, nil
		}

		if p.IsPage() {
			return false, nil
		}
		if currentBranchPrefix == "" || !strings.HasPrefix(ss, currentBranchPrefix) {
			if p.IsSection() && p.ShouldList(false) && p.Parent() == p {
				pages = append(pages, p)
			} else {
				w.SkipPrefix(ss + "/")
			}
		}
		currentBranchPrefix = ss + "/"
		return false, nil
	}

	if err := w.Walk(context.Background()); err != nil {
		panic(err)
	}

	valueobject.SortByDefault(pages)
	return pages
}

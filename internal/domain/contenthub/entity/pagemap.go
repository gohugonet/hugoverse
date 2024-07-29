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
	*ContentSpec

	// Main storage for all pages.
	*PageTrees

	Cache *valueobject.Cache

	PageBuilder *PageBuilder

	Log loggers.Logger
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

		//TODO in which dimension?
		m.TreePages.InsertWithLock(ps.Path().Base(), newPageTreesNode(p))

	}
	return nil
}

// Assemble
// Generalize this function
// - home page
// - section page
// - taxonomy page
// - term page
// - resource page
// - standalone page
// - Aggregates front matter, and mark changes
// - Clean page
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

	if err := m.assembleResources(); err != nil {
		return err
	}

	return nil
}

func (m *PageMap) assembleTerms() error {
	if err := m.PageBuilder.Term.Assemble(m.TreePages, m.PageBuilder); err != nil {
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

func (m *PageMap) assembleResources() error {
	pagesTree := m.PageTrees.TreePages // TODO: In which dimension
	resourcesTree := m.PageTrees.TreeResources

	lockType := doctree.LockTypeWrite
	w := &doctree.NodeShiftTreeWalker[*PageTreesNode]{
		Tree:     pagesTree,
		LockType: lockType,
		Handle: func(s string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {
			ps, found := n.getPage()
			if !found {
				return false, nil
			}

			if err := m.forEachResourceInPage(
				ps, lockType,
				func(resourceKey string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {
					rs, found := n.getResource()
					if !found {
						return false, nil
					}

					if !match.Has(doctree.DimensionLanguage) {
						// We got an alternative language version.
						// Clone this and insert it into the tree.
						rs = rs.clone()
						resourcesTree.InsertIntoCurrentDimension(resourceKey, rs)
					}
					if rs.r != nil {
						return false, nil
					}

					relPathOriginal := rs.Path().Unnormalized().PathRel(ps.Path().Unnormalized())
					relPath := rs.Path().BaseRel(ps.Path())

					var targetBasePaths []string
					if ps.s.Conf.IsMultihost() {
						baseTarget = targetPaths.SubResourceBaseLink
						// In multihost we need to publish to the lang sub folder.
						targetBasePaths = []string{ps.s.GetTargetLanguageBasePath()} // TODO(bep) we don't need this as a slice anymore.

					}

					rd := resources.ResourceSourceDescriptor{
						OpenReadSeekCloser:   rs.opener,
						Path:                 rs.path,
						GroupIdentity:        rs.path,
						TargetPath:           relPathOriginal, // Use the original path for the target path, so the links can be guessed.
						TargetBasePaths:      targetBasePaths,
						BasePathRelPermalink: targetPaths.SubResourceBaseLink,
						BasePathTargetPath:   baseTarget,
						NameNormalized:       relPath,
						NameOriginal:         relPathOriginal,
						LazyPublish:          !ps.m.pageConfig.Build.PublishResources,
					}
					r, err := ps.m.s.ResourceSpec.NewResource(rd)
					if err != nil {
						return false, err
					}
					rs.r = r
					return false, nil
				},
			); err != nil {
				return false, err
			}

			return false, nil
		},
	}

	return w.Walk(context.Background())
}

func (m *PageMap) forEachResourceInPage(
	ps contenthub.Page,
	lockType doctree.LockType,
	handle func(resourceKey string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error),
) error {
	keyPage := ps.Path().Path()
	if keyPage == "/" {
		keyPage = ""
	}
	prefix := paths.AddTrailingSlash(keyPage)
	isBranch := ps.Kind() != valueobject.KindPage

	rw := &doctree.NodeShiftTreeWalker[*PageTreesNode]{
		Tree:     m.TreeResources,
		Prefix:   prefix,
		LockType: lockType,
		Exact:    false,
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

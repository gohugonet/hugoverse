package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/doctree"
)

// PageTrees holds pages and resources in a tree structure for all sites/languages.
// Each site gets its own tree set via the Shape method.
type PageTrees struct {
	// This tree contains all Pages.
	// This include regular pages, sections, taxonomies and so on.
	// Note that all of these trees share the same key structure,
	// so you can take a leaf Page key and do a prefix search
	// with key + "/" to get all of its resources.
	TreePages *doctree.NodeShiftTree[contenthub.ContentNode]

	// This tree contains Resources bundled in pages.
	TreeResources *doctree.NodeShiftTree[contenthub.ContentNode]

	// All pages and resources.
	TreePagesResources doctree.WalkableTrees[contenthub.ContentNode]

	// This tree contains all taxonomy entries, e.g "/tags/blue/page1"
	TreeTaxonomyEntries *doctree.TreeShiftTree[contenthub.WeightedContentNode]

	// A slice of the resource trees.
	ResourceTrees doctree.MutableTrees
}

func (t *PageTrees) CreateMutableTrees() {
	t.treePagesResources = doctree.WalkableTrees[contentNodeI]{
		t.treePages,
		t.treeResources,
	}

	t.resourceTrees = doctree.MutableTrees{
		t.treeResources,
	}
}

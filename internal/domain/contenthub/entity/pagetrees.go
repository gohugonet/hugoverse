package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/pkg/cache/stale"
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
	TreePages *doctree.NodeShiftTree[*PageTreesNode]

	// This tree contains Resources bundled in pages.
	TreeResources *doctree.NodeShiftTree[*PageTreesNode]

	// All pages and resources.
	TreePagesResources doctree.WalkableTrees[*PageTreesNode]

	// This tree contains all taxonomy entries, e.g "/tags/blue/page1"
	TreeTaxonomyEntries *doctree.TreeShiftTree[*WeightedTermTreeNode]

	// A slice of the resource trees.
	ResourceTrees doctree.MutableTrees
}

func (t *PageTrees) CreateMutableTrees() {
	t.TreePagesResources = doctree.WalkableTrees[*PageTreesNode]{
		t.TreePages,
		t.TreeResources,
	}

	t.ResourceTrees = doctree.MutableTrees{
		t.TreeResources,
	}
}

type weightPage struct {
	*ordinalWeightPage
	page contenthub.Page
}

func (n *weightPage) Page() contenthub.Page {
	return n.page
}

func (n *weightPage) Owner() contenthub.Page {
	return n.ordinalWeightPage.Page
}

type ordinalWeightPage struct {
	ordinal int
	weight  int

	contenthub.Page
}

func (n *ordinalWeightPage) Weight() int {
	return n.weight
}

func (n *ordinalWeightPage) Ordinal() int {
	return n.ordinal
}

type WeightedTermTreeNode struct {
	*PageTreesNode
	term *ordinalWeightPage
}

type PageTreesNode struct {
	nodes map[contenthub.PageIdentity]contenthub.PageSource
}

func newPageTreesNode(ps contenthub.PageSource) *PageTreesNode {
	n := &PageTreesNode{
		nodes: make(map[contenthub.PageIdentity]contenthub.PageSource),
	}

	n.nodes[ps.PageIdentity()] = ps
	return n
}

func (n *PageTreesNode) merge(newNode *PageTreesNode) *PageTreesNode {
	// Create a map to track existing keys by their IDs
	existingKeys := make(map[string]contenthub.PageIdentity)
	for key := range n.nodes {
		existingKeys[key.PageLanguage()] = key
	}

	// Update or add entries from the new map
	for newKey, newValue := range newNode.nodes {
		if oldKey, exists := existingKeys[newKey.PageLanguage()]; exists {
			// Replace the old value with the new value
			n.nodes[oldKey] = newValue
		} else {
			// Add the new key-value pair to the old map
			n.nodes[newKey] = newValue
		}
	}
	return n
}

func (n *PageTreesNode) mergeWithLang(newNode *PageTreesNode, languageIndex int) *PageTreesNode {
	// Create a map to track existing keys by their IDs
	existingKeys := make(map[string]contenthub.PageIdentity)
	for key := range n.nodes {
		existingKeys[key.PageLanguage()] = key
	}

	// Update or add entries from the new map
	for newKey, newValue := range newNode.nodes {
		if oldKey, exists := existingKeys[newKey.PageLanguage()]; exists {
			if n.nodes[oldKey].PageIdentity().PageLanguageIndex() == languageIndex {
				_ = n.remove(oldKey)
			}
		}
		n.nodes[newKey] = newValue
	}
	return n
}

func (n *PageTreesNode) remove(k contenthub.PageIdentity) bool {
	v, exists := n.nodes[k]
	if !exists {
		return false
	}

	stale.MarkStale(v)
	delete(n.nodes, k)
	return true
}

func (n *PageTreesNode) delete(languageIndex int) bool {
	for k, _ := range n.nodes {
		if n.nodes[k].PageIdentity().PageLanguageIndex() == languageIndex {
			return n.remove(k)
		}
	}

	return false
}

func (n *PageTreesNode) isEmpty() bool {
	return len(n.nodes) == 0
}

func (n *PageTreesNode) shift(languageIndex int, exact bool) (*PageTreesNode, bool) {
	var firstV contenthub.PageSource = nil
	for k, v := range n.nodes {
		if firstV == nil {
			firstV = v
		}
		if n.nodes[k].PageIdentity().PageLanguageIndex() == languageIndex {
			return newPageTreesNode(v), true
		}
	}

	if firstV != nil && !exact {
		return newPageTreesNode(firstV), false
	}

	return nil, false
}

func (n *PageTreesNode) getPage() (contenthub.Page, bool) {
	for _, v := range n.nodes {
		return v.(contenthub.Page), true
	}
	return nil, false
}

func (n *PageTreesNode) getPages() []contenthub.Page {
	pages := make([]contenthub.Page, 0, len(n.nodes))
	for _, v := range n.nodes {
		pages = append(pages, v.(contenthub.Page))
	}
	return pages
}

func (n *PageTreesNode) getResource() (contenthub.PageSource, bool) {
	for _, v := range n.nodes {
		return v.(contenthub.PageSource), true
	}
	return nil, false
}

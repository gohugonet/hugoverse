package entity

import (
	"fmt"
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
	TreeResources *doctree.NodeShiftTree[*PageTreesNode]

	// All pages and resources.
	TreePagesResources doctree.WalkableTrees[contenthub.ContentNode]

	// This tree contains all taxonomy entries, e.g "/tags/blue/page1"
	TreeTaxonomyEntries *doctree.TreeShiftTree[contenthub.WeightedContentNode]

	// A slice of the resource trees.
	ResourceTrees doctree.MutableTrees
}

type PageTreesNode struct {
	nodes map[contenthub.PageIdentity]contenthub.PageSource
}

func newPageTreesNode(ps contenthub.PageSource) *PageTreesNode {
	n := &PageTreesNode{
		nodes: make(map[contenthub.PageIdentity]contenthub.PageSource),
	}

	n.nodes[ps.Identity()] = ps
	return n
}

func (n *PageTreesNode) merge(newNode *PageTreesNode) *PageTreesNode {
	// Create a map to track existing keys by their IDs
	existingKeys := make(map[string]contenthub.PageIdentity)
	for key := range n.nodes {
		existingKeys[key.IdentifierBase()] = key
	}

	// Update or add entries from the new map
	for newKey, newValue := range newNode.nodes {
		if oldKey, exists := existingKeys[newKey.IdentifierBase()]; exists {
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
		existingKeys[key.Language()] = key
	}

	// Update or add entries from the new map
	for newKey, newValue := range newNode.nodes {
		if oldKey, exists := existingKeys[newKey.Language()]; exists {
			// Replace the old value with the new value if language matches
			if n.nodes[oldKey].LanguageIndex() == languageIndex {
				delete(n.nodes, oldKey)
			}
		}
		n.nodes[newKey] = newValue
	}
	return n
}

func (n *PageTreesNode) shift(languageIndex int, exact bool) *PageTreesNode {
	for k, v := range n.nodes {
		if n.nodes[k].LanguageIndex() == languageIndex {
			return newPageTreesNode(v)
		}
	}

	if exact {
		fmt.Println("TODO exact for page resource, because page can share resource in different language")
	}

	return nil
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

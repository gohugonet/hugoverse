package entity

import "github.com/gohugonet/hugoverse/pkg/radixtree"

type ContentTree struct {
	Name string
	*radixtree.Tree
}

type ContentTrees []*ContentTree

type contentTreeNodeCallback func(s string, n *contentNode) bool

func (c ContentTrees) Walk(fn contentTreeNodeCallback) {
	for _, tree := range c {
		tree.Walk(func(s string, v any) bool {
			n := v.(*contentNode)
			return fn(s, n)
		})
	}
}

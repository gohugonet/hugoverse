package entity

import "github.com/gohugonet/hugoverse/pkg/doctree"

type Term struct {
	Terms map[string][]string
}

func (t *Term) Assemble(pages *doctree.NodeShiftTree[*PageTreesNode], pb *PageBuilder) error {
	return nil
}

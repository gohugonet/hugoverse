package valueobject

import "github.com/gohugonet/hugoverse/pkg/parser/pageparser"

type PageContentReplacement struct {
	Val []byte

	Source pageparser.Item
}

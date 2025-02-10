package valueobject

import "github.com/mdfriday/hugoverse/pkg/parser/pageparser"

type PageContentReplacement struct {
	Val []byte

	Source pageparser.Item
}

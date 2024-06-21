package valueobject

import "github.com/gohugonet/hugoverse/pkg/parser/pageparser"

type SourceParseInfo struct {
	// Items from the page parser.
	// These maps directly to the source
	ItemsStep1 pageparser.Items
}

func (s *SourceParseInfo) IsEmpty() bool {
	return len(s.ItemsStep1) == 0
}

package valueobject

import (
	"github.com/mdfriday/hugoverse/internal/domain/content"
)

type SortableContent []content.Sortable

func (s SortableContent) Len() int {
	return len(s)
}

func (s SortableContent) Less(i, j int) bool {
	return s[i].Time() > s[j].Time()
}

func (s SortableContent) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

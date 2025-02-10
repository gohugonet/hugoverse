package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/pkg/maps"
	"time"
)

const (
	Never       = "never"
	Always      = "always"
	ListLocally = "local"
	Link        = "link"
)

type Meta struct {
	List       string
	Parameters maps.Params
	Weight     int

	Date time.Time
}

func (m *Meta) Description() string {
	return ""
}

func (m *Meta) Params() maps.Params {
	return m.Parameters
}

func (m *Meta) Param(key any) (any, error) {
	return valueobject.Param(m, nil, key)
}

func (m *Meta) PageWeight() int {
	return m.Weight
}

func (m *Meta) PageDate() time.Time {
	return m.Date
}

func (m *Meta) PublishDate() time.Time {
	return m.PageDate()
}

// RelatedKeywords implements the related.Document interface needed for fast page searches.
func (m *Meta) RelatedKeywords(cfg contenthub.IndexConfig) ([]contenthub.Keyword, error) {
	v, err := m.Param(cfg.Name())
	if err != nil {
		return nil, err
	}

	return cfg.ToKeywords(v)
}

func (m *Meta) ShouldList(global bool) bool {
	return m.shouldList(global)
}

func (m *Meta) shouldList(global bool) bool {
	switch m.List {
	case Always:
		return true
	case Never:
		return false
	case ListLocally:
		return !global
	}
	return false
}

func (m *Meta) ShouldListAny() bool {
	return m.shouldListAny()
}

func (m *Meta) shouldListAny() bool {
	return m.shouldList(true) || m.shouldList(false)
}

func (m *Meta) NoLink() bool {
	return m.noLink()
}

func (m *Meta) noLink() bool {
	return false // TODO, updated based on configuration
}

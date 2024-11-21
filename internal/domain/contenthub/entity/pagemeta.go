package entity

import "github.com/gohugonet/hugoverse/pkg/maps"

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
}

func (m *Meta) Description() string {
	return ""
}

func (m *Meta) Params() maps.Params {
	return m.Parameters
}

func (m *Meta) PageWeight() int {
	return m.Weight
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

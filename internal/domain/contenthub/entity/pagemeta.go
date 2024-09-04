package entity

import "github.com/gohugonet/hugoverse/pkg/maps"

type Meta struct {
	// TODO
}

func (m *Meta) Description() string {
	return ""
}

func (m *Meta) Params() maps.Params {
	return maps.Params{}
}

func (m *Meta) Title() string {
	return ""
}

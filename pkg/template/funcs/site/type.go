package site

import "github.com/gohugonet/hugoverse/pkg/maps"

type Service interface {
	Author
	Meta
}

type Author interface {
	Name() string
	Email() string
}

type Meta interface {
	Params() maps.Params
}

package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/pkg/hexec"
)

type Security struct {
	valueobject.SecurityConfig
}

func (s Security) ExecAuth() hexec.ExecAuth {
	return s.Auth
}

package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/config/valueobject"
	"github.com/mdfriday/hugoverse/pkg/hexec"
)

type Security struct {
	valueobject.SecurityConfig
}

func (s Security) ExecAuth() hexec.ExecAuth {
	return s.Auth
}

func (s Security) CheckAllowedGetEnv(name string) error {
	if !s.Funcs.Getenv.Accept(name) {
		return &hexec.AccessDeniedError{
			Name:     name,
			Path:     "security.funcs.getenv",
			Policies: hexec.ToTOML(s.SecurityConfig),
		}
	}
	return nil
}

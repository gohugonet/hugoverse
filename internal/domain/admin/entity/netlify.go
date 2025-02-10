package entity

import "github.com/mdfriday/hugoverse/internal/domain/admin/valueobject"

type Netlify struct {
	Conf *valueobject.Config
}

func (a *Netlify) Token() string { return a.Conf.Netlify }

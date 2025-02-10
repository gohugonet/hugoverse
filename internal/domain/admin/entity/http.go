package entity

import "github.com/mdfriday/hugoverse/internal/domain/admin/valueobject"

type Http struct {
	Conf *valueobject.Config
}

func (a *Admin) Domain() string       { return a.Conf.Domain }
func (a *Admin) HttpPort() string     { return a.Conf.HTTPPort }
func (a *Admin) DevHttpsPort() string { return a.Conf.DevHTTPSPort }
func (a *Admin) BindAddress() string  { return a.Conf.BindAddress }

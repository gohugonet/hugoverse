package entity

import "github.com/gohugonet/hugoverse/internal/domain/admin/valueobject"

type Controller struct {
	Conf *valueobject.Config
}

func (a *Controller) CacheDisabled() bool { return a.Conf.DisableHTTPCache }
func (a *Controller) CorsDisabled() bool  { return a.Conf.DisableCORS }
func (a *Controller) GzipDisabled() bool  { return a.Conf.DisableGZIP }

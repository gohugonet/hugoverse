package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
	"net/url"
)

type Admin struct {
	Repo repository.Repository
	Conf *Config
}

func (a *Admin) SetConfig(data url.Values) error {
	return a.Repo.SetConfig(data)
}

func (a *Admin) PutConfig(key string, value any) error {
	err := a.Repo.PutConfig(key, value)
	if err != nil {
		return err
	}

	// check for "invalidate" value to reset the Etag
	if a.Conf.isCacheInvalidate() {
		a.RefreshETage()
	}

	return nil
}

func (a *Admin) InvalidateCache() error {
	err := a.PutConfig("etag", newEtag())
	if err != nil {
		return err
	}

	return nil
}

func (a *Admin) RefreshETage() {
	a.Conf.Etag = newEtag()
	a.Conf.CacheInvalidate = []string{}
}

func (a *Admin) Name() string         { return a.Conf.Name }
func (a *Admin) Domain() string       { return a.Conf.Domain }
func (a *Admin) ETage() string        { return a.Conf.Etag }
func (a *Admin) NewETage() string     { return newEtag() }
func (a *Admin) CacheMaxAge() int64   { return a.Conf.CacheMaxAge }
func (a *Admin) CacheDisabled() bool  { return a.Conf.DisableHTTPCache }
func (a *Admin) CorsDisabled() bool   { return a.Conf.DisableCORS }
func (a *Admin) GzipDisabled() bool   { return a.Conf.DisableGZIP }
func (a *Admin) ClientSecret() string { return a.Conf.ClientSecret }
func (a *Admin) HttpPort() int        { return a.Conf.HTTPPort }

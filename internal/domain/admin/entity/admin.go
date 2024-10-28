package entity

import (
	"encoding/json"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
	"github.com/gohugonet/hugoverse/internal/domain/admin/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"net/url"
)

type Admin struct {
	Repo repository.Repository
	Conf *valueobject.Config

	*Administrator
	*Upload
	*Http
	*Cache
	*Controller
	*Client

	Log loggers.Logger
}

func (a *Admin) ConfigEditor() ([]byte, error) {
	// TODO, remove it
	return a.Conf.MarshalEditor()
}

func (a *Admin) SetConfig(data url.Values) error {
	c, err := a.Conf.Convert(data)
	if err != nil {
		return err
	}

	m, err := c.Marshal()
	if err != nil {
		return err
	}

	if err := a.Repo.PutConfig(m); err != nil {
		return err
	}
	if err := a.LoadConfig(); err != nil {
		return err
	}
	// check for "invalidate" value to reset the Etag
	if a.Conf.IsCacheInvalidate() {
		a.RefreshETage()
	}

	return nil
}

func (a *Admin) LoadConfig() error {
	var conf *valueobject.Config

	data, err := a.Repo.LoadConfig()
	if err != nil {
		return err
	}

	if data == nil {
		conf = &valueobject.Config{}
	} else {
		err = json.Unmarshal(data, &conf)
		if err != nil {
			return err
		}
	}

	a.Conf = conf
	return nil
}

func (a *Admin) PutConfig(key string, value any) error {
	c, err := a.Conf.Update(key, value)
	if err != nil {
		return err
	}

	j, err := c.Marshal()
	if err != nil {
		return err
	}

	err = a.Repo.PutConfig(j)
	if err != nil {
		return err
	}

	a.Conf = c

	return nil
}

func (a *Admin) RefreshETage() {
	a.Conf.Etag = valueobject.NewEtag()
	a.Conf.CacheInvalidate = []string{}
}

func (a *Admin) InvalidateCache() error {
	err := a.PutConfig("etag", valueobject.NewEtag())
	if err != nil {
		return err
	}

	return nil
}

func (a *Admin) Name() string { return a.Conf.Name }

package entity

import (
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
	"github.com/gohugonet/hugoverse/internal/domain/content/factory"
	"github.com/gorilla/schema"
	"net/url"
)

type Admin struct {
	Repo repository.Repository
	Conf *Config
}

func (a *Admin) ConfigEditor() ([]byte, error) {
	// TODO, remove it
	return a.Conf.MarshalEditor()
}

func (a *Admin) NewUpload(data url.Values) error {
	var upload FileUpload

	decoder := schema.NewDecoder()
	if err := decoder.Decode(&upload, data); err != nil {
		return err
	}

	item, err := factory.NewItem()
	if err != nil {
		return err
	}
	upload.Item = *item

	slug, err := a.Repo.CheckUploadDuplication(upload.Name)
	if err != nil {
		return err
	}
	upload.Slug = slug

	nextId, err := a.Repo.NextUploadId()
	if err != nil {
		return err
	}
	upload.ID = int(nextId)

	uploadData, err := json.Marshal(upload)
	if err != nil {
		return err
	}

	return a.Repo.NewUpload(fmt.Sprintf("%d", upload.ID), slug, uploadData)
}

func (a *Admin) SetConfig(data url.Values) error {
	if err := a.Repo.SetConfig(data); err != nil {
		return err
	}
	if err := a.LoadConfig(); err != nil {
		return err
	}
	// check for "invalidate" value to reset the Etag
	if a.Conf.isCacheInvalidate() {
		a.RefreshETage()
	}

	return nil
}

func (a *Admin) LoadConfig() error {
	var conf *Config

	data, err := a.Repo.LoadConfig()
	if err != nil {
		return err
	}

	if data == nil {
		conf = &Config{}
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
	err := a.Repo.PutConfig(key, value)
	if err != nil {
		return err
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
func (a *Admin) HttpPort() string     { return a.Conf.HTTPPort }

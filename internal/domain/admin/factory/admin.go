package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/admin/entity"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/nilslice/jwt"
)

func NewAdminServer(db repository.Repository) (*entity.Admin, error) {
	a, err := NewAdmin(db)
	if err != nil {
		return nil, err
	}

	if a.ClientSecret() != "" {
		jwt.Secret([]byte(a.ClientSecret()))
	}
	err = a.InvalidateCache()
	if err != nil {
		return nil, err
	}

	return a, nil
}

func NewAdmin(repo repository.Repository) (*entity.Admin, error) {
	log := loggers.NewDefault()

	a := &entity.Admin{
		Repo: repo,

		Administrator: &entity.Administrator{
			Repo: repo,
			Log:  log,
		},
		Upload: &entity.Upload{
			Repo: repo,
		},
	}

	if err := a.LoadConfig(); err != nil {
		return nil, err
	}

	a.Http = &entity.Http{Conf: a.Conf}
	a.Cache = &entity.Cache{Conf: a.Conf}
	a.Controller = &entity.Controller{Conf: a.Conf}
	a.Client = &entity.Client{Conf: a.Conf}
	a.Netlify = &entity.Netlify{Conf: a.Conf}

	return a, nil
}

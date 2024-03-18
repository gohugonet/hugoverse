package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/admin"
	"github.com/gohugonet/hugoverse/internal/domain/admin/entity"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
)

func NewAdmin(repo repository.Repository) (admin.Admin, error) {
	a := &entity.Admin{
		Repo: repo,

		Administrator: &entity.Administrator{
			Repo: repo,
		},
		Upload: &entity.Upload{
			Repo: repo,
		},
	}

	if err := a.LoadConfig(); err != nil {
		return nil, err
	}
	a.Http = &entity.Http{
		Conf: a.Conf,
	}

	return a, nil
}

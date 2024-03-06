package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/admin"
	"github.com/gohugonet/hugoverse/internal/domain/admin/entity"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
)

func NewAdmin(repo repository.Repository) (admin.Admin, error) {
	a := &entity.Admin{
		Repo: repo,
	}
	if err := a.LoadConfig(); err != nil {
		return nil, err
	}
	return a, nil
}

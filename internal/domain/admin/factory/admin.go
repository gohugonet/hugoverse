package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/admin"
	"github.com/gohugonet/hugoverse/internal/domain/admin/entity"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
)

func NewAdmin(repo repository.Repository) admin.Admin {
	return &entity.Admin{Repo: repo}
}

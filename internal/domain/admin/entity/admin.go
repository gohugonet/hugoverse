package entity

import "github.com/gohugonet/hugoverse/internal/domain/admin/repository"

type Admin struct {
	Repo repository.Repository
}

func (a *Admin) PutConfig(key string, value any) error {
	return a.Repo.PutConfig(key, value)
}

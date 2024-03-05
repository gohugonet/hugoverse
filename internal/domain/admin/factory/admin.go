package factory

import (
	"encoding/json"
	"github.com/gohugonet/hugoverse/internal/domain/admin"
	"github.com/gohugonet/hugoverse/internal/domain/admin/entity"
	"github.com/gohugonet/hugoverse/internal/domain/admin/repository"
)

func NewAdmin(repo repository.Repository) (admin.Admin, error) {
	var conf *entity.Config

	data, err := repo.LoadConfig()
	if err != nil {
		return nil, err
	}

	if data == nil {
		conf = &entity.Config{}
	} else {
		err = json.Unmarshal(data, &conf)
		if err != nil {
			return nil, err
		}
	}

	return &entity.Admin{
		Repo: repo,
		Conf: conf,
	}, nil
}

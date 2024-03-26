package factory

import "github.com/gohugonet/hugoverse/internal/domain/site/entity"

func New() *entity.Site {
	return &entity.Site{}
}

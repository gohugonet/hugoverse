package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/entity"
	"github.com/gohugonet/hugoverse/internal/domain/content/repository"
)

func NewContent(repo repository.Repository) content.Content {
	c := &entity.Content{
		Types: make(map[string]func() interface{}),
		Repo:  repo,
	}
	c.Types["Demo"] = func() interface{} { return new(entity.Demo) }

	return c
}

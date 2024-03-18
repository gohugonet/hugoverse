package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/entity"
	"github.com/gohugonet/hugoverse/internal/domain/content/repository"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
)

func NewContent(repo repository.Repository) content.Content {
	c := &entity.Content{
		Types: make(map[string]content.Creator),
		Repo:  repo,
	}
	c.Types["Song"] = func() interface{} { return new(valueobject.Song) }
	c.Types["Student"] = func() interface{} { return new(valueobject.Student) }

	return c
}

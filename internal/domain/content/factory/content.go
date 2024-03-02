package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/entity"
)

func NewContent() content.Content {
	c := &entity.Content{Types: make(map[string]func() interface{})}
	c.Types["Demo"] = func() interface{} { return new(entity.Demo) }

	return c
}

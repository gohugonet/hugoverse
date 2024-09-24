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

	c.Types["Author"] = func() interface{} { return new(valueobject.Author) }
	c.Types["Language"] = func() interface{} { return new(valueobject.Language) }
	c.Types["Theme"] = func() interface{} { return new(valueobject.Theme) }
	c.Types["Post"] = func() interface{} { return new(valueobject.Post) }
	c.Types["Site"] = func() interface{} { return new(valueobject.Site) }
	c.Types["SiteLanguage"] = func() interface{} { return new(valueobject.SiteLanguage) }
	c.Types["SitePost"] = func() interface{} { return new(valueobject.SitePost) }

	return c
}

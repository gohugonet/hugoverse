package factory

import (
	"github.com/blevesearch/bleve"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/entity"
	"github.com/gohugonet/hugoverse/internal/domain/content/repository"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/spf13/afero"
)

func NewContent(repo repository.Repository, dir content.DirService) *entity.Content {
	log := loggers.NewDefault()
	log.Debugln("user data dir: ", repo.UserDataDir())

	c := &entity.Content{
		Types: make(map[string]content.Creator),
		Repo:  repo,

		Hugo: &entity.Hugo{
			Fs:         afero.NewOsFs(),
			DirService: dir,

			Log: log,
		},

		Log: log,
	}

	c.Types["Author"] = func() interface{} { return new(valueobject.Author) }
	c.Types["Language"] = func() interface{} { return new(valueobject.Language) }
	c.Types["Theme"] = func() interface{} { return new(valueobject.Theme) }
	c.Types["Post"] = func() interface{} { return new(valueobject.Post) }
	c.Types["Site"] = func() interface{} { return new(valueobject.Site) }
	c.Types["SiteLanguage"] = func() interface{} { return new(valueobject.SiteLanguage) }
	c.Types["SitePost"] = func() interface{} { return new(valueobject.SitePost) }

	c.Search = &entity.Search{
		ContentTypes: c.AllContentTypes(),
		Repo:         repo,
		Log:          log,

		IndicesMap: make(map[string]map[string]bleve.Index),
	}

	return c
}

func NewContentWithServices(repo repository.Repository, services content.Services, dirService content.DirService) *entity.Content {
	c := NewContent(repo, dirService)
	c.Hugo.Services = services

	return c
}

func NewItem() (*valueobject.Item, error) {
	return valueobject.NewItem()
}

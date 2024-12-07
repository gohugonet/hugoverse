package factory

import (
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
		UserTypes:  make(map[string]content.Creator),
		AdminTypes: make(map[string]content.Creator),
		Repo:       repo,

		Hugo: &entity.Hugo{
			Fs:         afero.NewOsFs(),
			DirService: dir,

			Log: log,
		},

		Log: log,
	}

	prepareUserTypes(c)
	prepareAdminTypes(c)

	c.Search = &entity.Search{
		TypeService: c,
		Repo:        repo,
		Log:         log,

		IndicesMap: make(map[string]*entity.CacheIndex),
	}

	return c
}

func prepareUserTypes(c *entity.Content) {
	c.UserTypes["Author"] = func() interface{} { return new(valueobject.Author) }
	c.UserTypes["Language"] = func() interface{} { return new(valueobject.Language) }
	c.UserTypes["Theme"] = func() interface{} { return new(valueobject.Theme) }
	c.UserTypes["Post"] = func() interface{} { return new(valueobject.Post) }
	c.UserTypes["Resource"] = func() interface{} { return new(valueobject.Resource) }
	c.UserTypes["Site"] = func() interface{} { return new(valueobject.Site) }
	c.UserTypes["SiteLanguage"] = func() interface{} { return new(valueobject.SiteLanguage) }
	c.UserTypes["SitePost"] = func() interface{} { return new(valueobject.SitePost) }
	c.UserTypes["SiteResource"] = func() interface{} { return new(valueobject.SiteResource) }
	c.UserTypes["Deployment"] = func() interface{} { return new(valueobject.Deployment) }
}

func prepareAdminTypes(c *entity.Content) {
	c.AdminTypes["Domain"] = func() interface{} { return new(valueobject.Domain) }
}

func NewContentWithServices(repo repository.Repository, services content.Services, dirService content.DirService) *entity.Content {
	c := NewContent(repo, dirService)
	c.Hugo.Services = services

	return c
}

func NewItem() (*valueobject.Item, error) {
	return valueobject.NewItem()
}

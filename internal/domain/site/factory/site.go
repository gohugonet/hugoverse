package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/entity"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

func New(services site.Services) *entity.Site {
	s := &entity.Site{
		ContentSvc:   services,
		ResourcesSvc: services,

		Template: nil,

		Publisher: &entity.Publisher{Fs: services.Publish()},

		Title:    "Hugoverse",                                              // TODO: Make configurable
		Author:   valueobject.NewAuthor("Hugoverse", "support@gohugo.net"), // TODO: Make configurable
		Compiler: valueobject.NewVersion("0.0.0"),                          // TODO: Make configurable

		URL: &entity.URL{
			Base:      services.BaseUrl(),
			Canonical: true,
		},
		Language: &entity.Language{
			LangSvc: services,
		},
		Navigation: &entity.Navigation{},

		Log: loggers.NewDefault(),
	}

	s.PrepareLazyLoads()

	return s
}

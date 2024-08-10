package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/entity"
	"github.com/gohugonet/hugoverse/pkg/loggers"
)

func New(services site.Services) *entity.Site {
	return &entity.Site{
		ContentSvc:   services,
		ResourcesSvc: services,

		Template: nil,

		Publisher: &entity.Publisher{Fs: services.Publish()},

		URL: &entity.URL{
			Base:      services.BaseUrl(),
			Canonical: true,
		},
		Language: &entity.Language{
			LangSvc: services,
		},

		Log: loggers.NewDefault(),
	}
}

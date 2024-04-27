package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/entity"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/media"
)

func New(fs site.Fs, ch contenthub.ContentHub, conf site.Config) site.Site {
	mediaTypes := media.DecodeTypes()
	formats := valueobject.DecodeFormats(mediaTypes)
	outputFormats := valueobject.CreateSiteOutputFormats(formats)

	return &entity.Site{
		OutputFormats:       outputFormats,
		OutputFormatsConfig: formats,
		MediaTypesConfig:    mediaTypes,

		Publisher: &entity.DestinationPublisher{Fs: fs.Publish()},

		ContentHub: ch,

		URL: &entity.URL{
			Base:      conf.BaseUrl(),
			Canonical: true,
		},
		Language: &entity.Language{
			Config: conf.Languages(),
		},
	}
}

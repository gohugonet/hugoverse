package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/entity"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
)

func New(fs site.Fs, cs site.ContentSpec) *entity.Site {
	mediaTypes := valueobject.DecodeTypes()
	formats := valueobject.DecodeFormats(mediaTypes)
	outputFormats := valueobject.CreateSiteOutputFormats(formats)

	return &entity.Site{
		OutputFormats:       outputFormats,
		OutputFormatsConfig: formats,
		MediaTypesConfig:    mediaTypes,

		Publisher: &entity.DestinationPublisher{Fs: fs.Publish()},

		ContentSpec: cs,
	}
}

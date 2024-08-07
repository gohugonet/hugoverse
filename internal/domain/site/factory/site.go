package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/entity"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/output"
)

func New(services site.Services) *entity.Site {
	mediaTypes := media.DecodeTypes()
	formats := output.DecodeFormats(mediaTypes)
	outputFormats := CreateSiteOutputFormats(formats)

	return &entity.Site{
		OutputFormats:       outputFormats,
		OutputFormatsConfig: formats,
		MediaTypesConfig:    mediaTypes,

		Publisher: &entity.DestinationPublisher{Fs: services.Publish()},

		ContentSvc: services,
		Template:   nil,

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

func CreateSiteOutputFormats(allFormats output.Formats) map[string]output.Formats {
	defaultOutputFormats :=
		createDefaultOutputFormats(allFormats)
	return defaultOutputFormats
}

func createDefaultOutputFormats(allFormats output.Formats) map[string]output.Formats {
	htmlOut, _ := allFormats.GetByName(output.HTMLFormat.Name)

	m := map[string]output.Formats{
		contenthub.KindPage:    {htmlOut},
		contenthub.KindHome:    {htmlOut},
		contenthub.KindSection: {htmlOut},
	}

	return m
}

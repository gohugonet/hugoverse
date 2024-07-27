package entity

import (
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/output"
	"sort"
	"time"
)

type Site struct {
	// Output formats defined in site config per Page Kind, or some defaults
	// if not set.
	// Output formats defined in Page front matter will override these.
	OutputFormats map[string]output.Formats

	// The output formats that we need to render this site in. This slice
	// will be fixed once set.
	// This will be the union of Site.Pages' outputFormats.
	// This slice will be sorted.
	RenderFormats output.Formats

	// All the output formats and media types available for this site.
	// These values will be merged from the Hugo defaults, the site config and,
	// finally, the language settings.
	OutputFormatsConfig output.Formats
	MediaTypesConfig    media.Types

	Publisher site.Publisher

	ContentHub contenthub.ContentHub

	Template site.Template

	*URL
	*Language

	Log     loggers.Logger `json:"-"`
	siteLog logg.LevelLogger
}

func (s *Site) Build(t site.Template) error {
	s.siteLog = s.Log.InfoCommand("site build")
	defer loggers.TimeTrackf(s.siteLog, time.Now(), nil, "")

	s.Template = t

	if err := s.setup(); err != nil {
		return err
	}
	for _, l := range s.Language.Config {
		s.Language.currentLanguage = l
		err := s.render()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Site) setup() error {
	l := s.siteLog.WithField("step setup", "setup url and languages")
	start := time.Now()
	defer func() {
		loggers.TimeTrackf(l, start, nil, "")
	}()

	if err := s.Template.MarkReady(); err != nil {
		return err
	}

	if err := s.URL.setup(); err != nil {
		return err
	}
	if err := s.Language.setup(); err != nil {
		return err
	}
	return nil
}

func (s *Site) render() error {
	l := s.siteLog.WithField("step render", "render sties")
	start := time.Now()
	defer func() {
		loggers.TimeTrackf(l, start, nil, "")
	}()

	s.initRenderFormats()

	// Get page output ready
	if err := s.preparePagesForRender(); err != nil {
		return err
	}
	if err := s.renderPages(); err != nil {
		return err
	}

	return nil
}

func (s *Site) renderPages() error {
	for _, of := range s.RenderFormats {
		td := valueobject.NewTemplateDescriptor(of.Name, of.MediaType.SubType)

		err := s.ContentHub.RenderPages(td, func(info contenthub.PageInfo) error {
			pp, err := valueobject.NewPagePaths(s.OutputFormatsConfig, info)
			if err != nil {
				return err
			}

			pd := site.Descriptor{
				Src:          info.Buffer(),
				TargetPath:   pp.TargetPaths[of.Name].Paths.TargetFilename,
				OutputFormat: of,
			}
			return s.Publisher.Publish(pd)
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Site) preparePagesForRender() error {
	return s.ContentHub.PreparePages()
}

func (s *Site) initRenderFormats() {
	formatSet := make(map[string]bool)
	formats := output.Formats{}

	// media type - format
	// site output format - render format
	// Add the per kind configured output formats
	for _, kind := range contenthub.AllKindsInPages {
		if siteFormats, found := s.OutputFormats[kind]; found {
			for _, f := range siteFormats {
				if !formatSet[f.Name] {
					formats = append(formats, f)
					formatSet[f.Name] = true
				}
			}
		}
	}

	sort.Sort(formats)

	// HTML
	s.RenderFormats = formats
}

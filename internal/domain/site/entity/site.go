package entity

import (
	"bytes"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"sort"
)

type Site struct {
	// Output formats defined in site config per Page Kind, or some defaults
	// if not set.
	// Output formats defined in Page front matter will override these.
	OutputFormats map[string]valueobject.Formats

	// The output formats that we need to render this site in. This slice
	// will be fixed once set.
	// This will be the union of Site.Pages' outputFormats.
	// This slice will be sorted.
	RenderFormats valueobject.Formats

	// All the output formats and media types available for this site.
	// These values will be merged from the Hugo defaults, the site config and,
	// finally, the language settings.
	OutputFormatsConfig valueobject.Formats
	MediaTypesConfig    valueobject.Types

	Publisher site.Publisher

	ContentSpec site.ContentSpec
}

func (s *Site) Build() error {
	err := s.render()
	if err != nil {
		return err
	}

	return nil
}

func (s *Site) render() error {
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
	return s.ContentSpec.RenderPages(func(kind string, sec []string, dir, name string, buf *bytes.Buffer) error {
		pp, err := valueobject.NewPagePaths(s.OutputFormatsConfig, kind, sec, dir, name)
		if err != nil {
			return err
		}

		for _, of := range s.RenderFormats {
			pd := site.Descriptor{
				Src:          buf,
				TargetPath:   pp.TargetPaths[of.Name].Paths.TargetFilename,
				OutputFormat: of,
			}
			return s.Publisher.Publish(pd)
		}

		return nil
	})
}

func (s *Site) preparePagesForRender() error {
	return s.ContentSpec.PreparePages()
}

func (s *Site) initRenderFormats() {
	formatSet := make(map[string]bool)
	formats := valueobject.Formats{}

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

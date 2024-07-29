package entity

import (
	"fmt"
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/pkg/env"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/output"
	"sort"
	"sync"
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

	ContentSvc site.ContentService

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
	for _, l := range s.LangSvc.LanguageKeys() {
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

	spc.clear()

	if err := s.renderPages(); err != nil {
		return err
	}

	return nil
}

func (s *Site) renderPages() error {
	numWorkers := env.GetNumWorkerMultiplier()

	results := make(chan error)
	pages := make(chan *Page, numWorkers) // buffered for performance
	errs := make(chan error)

	go s.errorCollator(results, errs)

	wg := &sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go pageRenderer(s, pages, results, wg)
	}

	if err := s.ContentSvc.WalkPages(s.Language.CurrentLanguageIndex(), func(p contenthub.Page) error {
		pages <- &Page{Page: p}

		return nil
	}); err != nil {
		close(pages)
		close(results)

		return fmt.Errorf("failed to walk pages: %w", herrors.ImproveIfNilPointer(err))
	}

	close(pages)

	wg.Wait()

	close(results)

	err := <-errs
	if err != nil {
		return fmt.Errorf("failed to render pages: %w", herrors.ImproveIfNilPointer(err))
	}
	return nil
}

func (s *Site) errorCollator(results <-chan error, errs chan<- error) {
	var errors []error
	for e := range results {
		errors = append(errors, e)
	}

	errs <- s.pickOneAndLogTheRest(errors)

	close(errs)
}

func (s *Site) pickOneAndLogTheRest(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	var i int

	for j, err := range errors {
		// If this is in server mode, we want to return an error to the client
		// with a file context, if possible.
		if herrors.UnwrapFileError(err) != nil {
			i = j
			break
		}
	}

	// Log the rest, but add a threshold to avoid flooding the log.
	const errLogThreshold = 5

	for j, err := range errors {
		if j == i || err == nil {
			continue
		}

		if j >= errLogThreshold {
			break
		}

		s.Log.Errorln(err)
	}

	return errors[i]
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

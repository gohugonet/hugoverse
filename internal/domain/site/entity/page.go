package entity

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"path"
	"sync"
)

type WeightedPage struct {
	*Page
	contenthub.OrdinalWeightPage
}

type Page struct {
	resSvc  site.ResourceService
	tmplSvc site.Template
	langSvc site.LanguageService

	publisher *Publisher
	git       *valueobject.GitMap

	contenthub.Page
	contenthub.PageOutput

	*Site

	resources []resources.Resource

	dataInit sync.Once
	data     Data
}

func (p *Page) processResources(pageSources []contenthub.PageSource) error {
	for _, source := range pageSources {
		rs, err := p.resSvc.GetResourceWithOpener(source.Paths().Path(), source.Opener())
		if err != nil {
			return err
		}
		p.resources = append(p.resources, rs)
	}

	return nil
}

func (p *Page) render() error {
	if err := p.renderResources(); err != nil {
		return err
	}

	if err := p.renderPage(); err != nil {
		return err
	}

	return nil
}

func (p *Page) renderPage() error {
	layouts := p.Layouts()
	tmpl, found, err := p.tmplSvc.LookupLayout(layouts)
	if err != nil {
		return err
	}
	if !found {
		p.Log.Warnf("failed to find layout: %s, for page %s", layouts, p.Paths().Path())

		return nil
	}

	renderBuffer := bp.GetBuffer()
	defer bp.PutBuffer(renderBuffer)

	outputs, err := p.PageOutputs()
	if err != nil {
		return p.errorf(err, "failed to get page outputs")
	}

	for _, o := range outputs {
		p.PageOutput = o

		var targetFilenames []string

		prefix := o.TargetPrefix()
		if p.Site.currentLanguage == prefix && prefix == p.LanguageSvc.DefaultLanguage() {
			prefix = ""
		} else {
			prefix = p.Site.currentLanguage
		}

		targetFilenames = append(targetFilenames, path.Join(prefix, o.TargetFilePath()))

		if err := p.renderAndWritePage(tmpl, renderBuffer, targetFilenames); err != nil {
			return err
		}

		if p.Current() != nil {
			for current := p.Current().Next(); current != nil; current = current.Next() {
				p.SetCurrent(current)

				targetFilenames = []string{path.Join(prefix, current.URL(), o.TargetFileBase())}
				if err := p.renderAndWritePage(tmpl, renderBuffer, targetFilenames); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (p *Page) renderAndWritePage(tmpl template.Preparer, renderBuffer *bytes.Buffer, targetFilenames []string) error {
	if err := p.tmplSvc.ExecuteWithContext(context.Background(), tmpl, renderBuffer, p); err != nil {
		p.Log.Errorf("failed to execute template: %s", err)
		return err
	}
	if err := p.publisher.PublishSource(renderBuffer, targetFilenames...); err != nil {
		return p.errorf(err, "failed to publish page")
	}
	renderBuffer.Reset()

	return nil
}

func (p *Page) renderResources() error {
	for _, rs := range p.resources {
		var targetFilenames []string

		outputs, err := p.PageOutputs()
		if err != nil {
			return p.errorf(err, "failed to get page outputs")
		}
		for _, o := range outputs {
			prefix := o.TargetPrefix()
			if p.Site.currentLanguage == prefix && prefix == p.LanguageSvc.DefaultLanguage() {
				prefix = ""
			} else {
				prefix = p.Site.currentLanguage
			}

			targetFilenames = append(targetFilenames, path.Join(prefix, rs.TargetPath()))
		}

		if err := func() error {
			fr, err := rs.ReadSeekCloser()
			defer fr.Close()

			if err != nil {
				return p.errorf(err, "failed to open resource for reading")
			}

			if err := p.publisher.PublishFiles(fr, targetFilenames...); err != nil {
				return p.errorf(err, "failed to publish page resources")
			}

			return nil

		}(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Page) errorf(err error, format string, a ...any) error {
	if herrors.UnwrapFileError(err) != nil {
		// More isn't always better.
		return err
	}
	args := append([]any{p.PageIdentity().PageLanguage(), p.Paths().Path()}, a...)
	args = append(args, err)
	format = "[%s] page %q: " + format + ": %w"
	if err == nil {
		return fmt.Errorf(format, args...)
	}
	return fmt.Errorf(format, args...)
}

func (p *Page) clone() *Page {
	np := *p

	return &np
}

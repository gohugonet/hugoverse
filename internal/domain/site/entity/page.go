package entity

import (
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"path"
)

type Pages []Page

// Len returns the number of pages in the list.
func (p Pages) Len() int {
	return len(p)
}
func (ps Pages) String() string {
	return fmt.Sprintf("Pages(%d)", len(ps))
}

type Page struct {
	resSvc  site.ResourceService
	tmplSvc site.Template

	publisher *Publisher

	contenthub.Page

	resources []resources.Resource
}

func (p *Page) processResources(pageSources []contenthub.PageSource) error {
	for _, source := range pageSources {
		rs, err := p.resSvc.GetResourceWithOpener(source.Path().Path(), source.Opener())
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
		return fmt.Errorf("failed to find layout: %s, for page %s", layouts, p.Path().Path())
	}

	renderBuffer := bp.GetBuffer()
	defer bp.PutBuffer(renderBuffer)

	if err := p.tmplSvc.ExecuteWithContext(context.Background(), tmpl, renderBuffer, p); err != nil {
		return err
	}

	var targetFilenames []string

	outputs, err := p.PageOutputs()
	if err != nil {
		return p.errorf(err, "failed to get page outputs")
	}
	for _, o := range outputs {
		targetFilenames = append(targetFilenames, path.Join(o.TargetPrefix(), o.TargetFilePath()))
	}

	if err := p.publisher.PublishSource(renderBuffer, targetFilenames...); err != nil {
		return p.errorf(err, "failed to publish page")
	}

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
			targetFilenames = append(targetFilenames, path.Join(o.TargetPrefix(), o.TargetSubResourceDir(), rs.TargetPath()))
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
	args := append([]any{p.Language(), p.Path().Path()}, a...)
	args = append(args, err)
	format = "[%s] page %q: " + format + ": %w"
	if err == nil {
		return fmt.Errorf(format, args...)
	}
	return fmt.Errorf(format, args...)
}

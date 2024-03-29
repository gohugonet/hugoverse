package entity

import (
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/pkg/bufferpool"
	"io"
	"sync"
)

type render struct {
	pageMap          *PageMap
	templateExecutor contenthub.TemplateExecutor
	td               contenthub.TemplateDescriptor
	cb               func(info contenthub.PageInfo) error
}

// renderPages renders pages each corresponding to a markdown file.
func (ch *render) renderPages() error {
	numWorkers := 3

	results := make(chan error)
	pages := make(chan *pageState, numWorkers) // buffered for performance
	errs := make(chan error)

	go ch.errorCollator(results, errs)

	wg := &sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go ch.pageRenderer(pages, results, wg)
	}

	var count int
	ch.pageMap.PageTrees.Walk(func(ss string, n *contentNode) bool {
		select {
		default:
			count++
			fmt.Println("777 count: ", count, ss, n, n.p)
			pages <- n.p
		}

		return false
	})

	close(pages)

	wg.Wait()

	close(results)

	err := <-errs
	if err != nil {
		return fmt.Errorf("failed to render pages: %w", err)
	}
	return nil
}

func (ch *render) pageRenderer(pages <-chan *pageState, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for p := range pages {
		fmt.Printf(">>>> page: %#+v\n", p)

		templ, found, err := ch.resolveTemplate(p)
		if err != nil {
			fmt.Println("failed to resolve template")
			continue
		}

		if !found { // layout: "", kind: section, name: HTML
			fmt.Printf("layout: %s, kind: %s", p.Layout(), p.Kind())
			continue
		}

		if err := ch.renderAndWritePage(p, templ); err != nil {
			fmt.Println(" render err")
			fmt.Printf("%#v", err)
			results <- err
		}
	}
}

func (ch *render) renderAndWritePage(p *pageState, templ template.Template) error {
	renderBuffer := bufferpool.GetBuffer()
	defer bufferpool.PutBuffer(renderBuffer)

	if err := ch.renderForTemplate(p.Kind(), p, renderBuffer, templ); err != nil {
		return err
	}

	if renderBuffer.Len() == 0 {
		return nil
	}

	var (
		dir      string
		baseName string
	)

	if !p.File().IsZero() {
		dir = p.File().Dir()
		baseName = p.File().TranslationBaseName()
	}

	return ch.cb(valueobject.NewPageInfo(baseName, p.Kind(), dir, p.SectionsEntries(), renderBuffer))
}

func (ch *render) renderForTemplate(name string, d any, w io.Writer, templ template.Template) (err error) {
	if templ == nil {
		fmt.Println("templ is nil")
		return nil
	}

	if err = ch.templateExecutor.ExecuteWithContext(context.Background(), templ, w, d); err != nil {
		return fmt.Errorf("render of %q failed: %w", name, err)
	}
	return
}

func (ch *render) resolveTemplate(p *pageState) (template.Template, bool, error) {
	d := p.getLayoutDescriptor(ch.td)

	lh := valueobject.NewLayoutHandler()
	names, err := lh.For(d)
	if err != nil {
		return nil, false, err
	}

	return ch.templateExecutor.LookupLayout(names)
}

func (ch *render) errorCollator(results <-chan error, errs chan<- error) {
	var errors []error
	for e := range results {
		errors = append(errors, e)
	}

	if len(errors) > 0 {
		errs <- fmt.Errorf("failed to render pages: %v", errors)
	}

	close(errs)
}

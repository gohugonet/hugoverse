package entity

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/pkg/bufferpool"
	"io"
	"sync"
)

var contentSpec *ContentSpec

type ContentHub struct {
	Fs contenthub.Fs

	// ExecTemplate handling.
	TemplateExecutor contenthub.TemplateExecutor

	*ContentSpec

	*PageCollections

	cb func(kind string, sec []string, dir, name string, buf *bytes.Buffer) error
}

func (ch *ContentHub) CS() {
	contentSpec = ch.ContentSpec
}

func (ch *ContentHub) CollectPages() error {
	if err := ch.process(); err != nil {
		return fmt.Errorf("process: %w", err)
	}

	if err := ch.assemble(); err != nil {
		return fmt.Errorf("assemble: %w", err)
	}

	return nil
}

func (ch *ContentHub) process() error {
	if err := ch.readAndProcessContent(); err != nil {
		return fmt.Errorf("readAndProcessContent: %w", err)
	}
	return nil
}

func (ch *ContentHub) readAndProcessContent() error {
	proc := newPagesProcessor(ch.PageCollections.PageMap)
	c := newPagesCollector(proc, ch.Fs.ContentFs())
	if err := c.Collect(); err != nil {
		return err
	}

	return nil
}

func (ch *ContentHub) assemble() error {
	if err := ch.PageCollections.PageMap.Assemble(); err != nil {
		return err
	}
	return nil
}

func (ch *ContentHub) PreparePages() error {
	var err error
	ch.PageCollections.PageMap.withEveryBundlePage(func(p *pageState) bool {
		if err = p.initOutputFormat(); err != nil {
			return true
		}
		return false
	})
	return nil
}

func (ch *ContentHub) RenderPages(
	cb func(kind string, sec []string, dir, name string, buf *bytes.Buffer) error) error {
	ch.cb = cb

	if err := ch.renderPages(); err != nil {
		return fmt.Errorf("renderPages: %w", err)
	}

	return nil
}

// renderPages renders pages each corresponding to a markdown file.
func (ch *ContentHub) renderPages() error {
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
	ch.PageCollections.PageMap.PageTrees.Walk(func(ss string, n *contentNode) bool {
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

func (ch *ContentHub) pageRenderer(pages <-chan *pageState, results chan<- error, wg *sync.WaitGroup) {
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

func (ch *ContentHub) renderAndWritePage(p *pageState, templ template.Template) error {
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

	return ch.cb(
		p.Kind(),
		p.SectionsEntries(),
		dir,
		baseName,
		renderBuffer)
}

func (ch *ContentHub) renderForTemplate(name string, d any, w io.Writer, templ template.Template) (err error) {
	if templ == nil {
		fmt.Println("templ is nil")
		return nil
	}

	if err = ch.TemplateExecutor.ExecuteWithContext(context.Background(), templ, w, d); err != nil {
		return fmt.Errorf("render of %q failed: %w", name, err)
	}
	return
}

func (ch *ContentHub) resolveTemplate(p *pageState) (template.Template, bool, error) {
	f := valueobject.HTMLFormat // set in shiftToOutputFormat
	d := p.getLayoutDescriptor()

	lh := valueobject.NewLayoutHandler()
	names, err := lh.For(d, f)
	if err != nil {
		return nil, false, err
	}

	return ch.TemplateExecutor.LookupLayout(names)
}

func (ch *ContentHub) errorCollator(results <-chan error, errs chan<- error) {
	var errors []error
	for e := range results {
		errors = append(errors, e)
	}

	if len(errors) > 0 {
		errs <- fmt.Errorf("failed to render pages: %v", errors)
	}

	close(errs)
}

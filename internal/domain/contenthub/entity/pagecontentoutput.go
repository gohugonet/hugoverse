package entity

import (
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/markdown"
	"github.com/gohugonet/hugoverse/pkg/lazy"
	"html/template"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

func newPageContentOutput(p *pageState) (*pageContentOutput, error) {
	parent := p.init

	cp := &pageContentOutput{
		p:           p,
		renderHooks: &renderHooks{},
	}

	initContent := func() (err error) {
		if p.cmap == nil {
			// Nothing to do.
			return nil
		}
		defer func() {
			// See https://github.com/gohugonet/hugoverse/pkgissues/6210
			if r := recover(); r != nil {
				err = fmt.Errorf("%s", r)
				fmt.Printf("[BUG] Got panic:\n%s\n%s", r, string(debug.Stack()))
			}
		}()

		cp.workContent = p.contentToRender(p.source.parsed, p.cmap)

		r, err := cp.renderContent(cp.workContent, true)
		if err != nil {
			return err
		}

		cp.workContent = r.Bytes()
		cp.content = BytesToHTML(cp.workContent)

		return nil
	}

	// There may be recursive loops in shortcodes and render hooks.
	cp.initMain = parent.BranchWithTimeout(30*time.Second, func(ctx context.Context) (any, error) {
		return nil, initContent()
	})

	cp.initPlain = cp.initMain.Branch(func() (any, error) {
		cp.plain = string(cp.content)
		cp.plainWords = strings.Fields(cp.plain)

		return nil, nil
	})

	return cp, nil
}

// BytesToHTML converts bytes to type template.HTML.
func BytesToHTML(b []byte) template.HTML {
	return template.HTML(string(b))
}

// pageContentOutput represents the Page content for a given output format.
type pageContentOutput struct {
	p *pageState

	// Lazy load dependencies
	initMain  *lazy.Init
	initPlain *lazy.Init

	workContent []byte

	// Content sections
	content template.HTML
	summary template.HTML

	plainWords []string
	plain      string

	// Renders Markdown hooks.
	renderHooks *renderHooks
}

type renderHooks struct {
	getRenderer markdown.GetRendererFunc
	init        sync.Once
}

func (cp *pageContentOutput) renderContent(content []byte, renderTOC bool) (contenthub.Result, error) {
	if err := cp.initRenderHooks(); err != nil {
		return nil, err
	}

	c := cp.p.getContentConverter()
	return cp.renderContentWithConverter(c, content, renderTOC)
}

func (cp *pageContentOutput) initRenderHooks() error {
	if cp == nil {
		return nil
	}

	cp.renderHooks.init.Do(func() {
		// TODO, no template based rendering, use markdown default render
		cp.renderHooks.getRenderer = markdown.DefaultRendererFunc
	})

	return nil
}

func (cp *pageContentOutput) renderContentWithConverter(
	c contenthub.Converter, content []byte, renderTOC bool) (contenthub.Result, error) {

	fmt.Println("renderContentWithConverter", string(content), renderTOC)

	r, err := c.Convert(
		markdown.RenderContext{
			Ctx:         context.Background(),
			Src:         content,
			RenderTOC:   renderTOC,
			GetRenderer: cp.renderHooks.getRenderer,
		})

	return r, err
}

func (cp *pageContentOutput) Content() (any, error) {
	if cp.initInit(cp.initMain) {
		return cp.content, nil
	}
	return nil, nil
}

func (cp *pageContentOutput) initInit(init *lazy.Init) bool {
	_, err := init.Do()
	if err != nil {
		fmt.Printf("fatal error %v", err)
	}
	return err == nil
}

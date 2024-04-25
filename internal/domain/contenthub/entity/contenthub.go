package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
)

type ContentHub struct {
	Fs contenthub.Fs

	// ExecTemplate handling.
	TemplateExecutor contenthub.TemplateExecutor

	*PageCollections

	*render
}

func (ch *ContentHub) SetTemplateExecutor(exec contenthub.TemplateExecutor) {
	ch.TemplateExecutor = exec
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

func (ch *ContentHub) RenderPages(td contenthub.TemplateDescriptor, cb func(info contenthub.PageInfo) error) error {
	ch.render = &render{
		pageMap:          ch.PageCollections.PageMap,
		templateExecutor: ch.TemplateExecutor,
		td:               td,
		cb:               cb,
	}

	if err := ch.renderPages(); err != nil {
		return fmt.Errorf("renderPages: %w", err)
	}

	return nil
}

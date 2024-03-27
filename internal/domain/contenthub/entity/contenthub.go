package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
)

type ContentHub struct {
	Fs contenthub.Fs

	// ExecTemplate handling.
	TemplateExecutor contenthub.TemplateExecutor

	*ContentSpec

	*PageCollections
}

func (ch *ContentHub) Process() error {
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

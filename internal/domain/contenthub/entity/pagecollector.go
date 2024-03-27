package entity

import (
	"context"
	"fmt"
	fsFactory "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/spf13/afero"
)

func newPagesCollector(proc pagesCollectorProcessorProvider, fs afero.Fs) *pagesCollector {
	return &pagesCollector{
		proc: proc,
		fs:   fs,
	}
}

type pagesCollector struct {
	fs   afero.Fs
	proc pagesCollectorProcessorProvider
}

// Collect pages.
func (c *pagesCollector) Collect() (collectErr error) {
	c.proc.Start(context.Background())
	defer func() {
		err := c.proc.Wait()
		if collectErr == nil {
			collectErr = err
		}
	}()

	collectErr = c.collectDir("")
	return
}

func (c *pagesCollector) collectDir(dirname string) error {
	w := fsFactory.NewWalkway(c.fs, dirname, func(path string, info fsVO.FileMetaInfo, err error) error {
		if err := c.handleFile(info); err != nil {
			fmt.Println("collectDir --- ", path, err)
			return err
		}
		return nil
	})

	return w.Walk()
}

func (c *pagesCollector) handleFile(fi fsVO.FileMetaInfo) error {
	if fi.IsDir() {
		return nil
	}

	if err := c.proc.Process(fi); err != nil {
		return err
	}

	return nil
}

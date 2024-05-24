package entity

import (
	"context"
	"fmt"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/pkg/log"
	"golang.org/x/sync/errgroup"
)

func newPagesProcessor(pm *PageMap) *pagesProcessor {
	return &pagesProcessor{
		processor: &sitePagesProcessor{
			m:        pm,
			itemChan: make(chan interface{}, 1),
			log:      log.NewStdLogger(),
		},
	}
}

type pagesCollectorProcessorProvider interface {
	Process(item any) error
	Start(ctx context.Context) context.Context
	Wait() error
}

type pagesProcessor struct {
	processor pagesCollectorProcessorProvider
}

func (proc *pagesProcessor) Process(item any) error {
	switch v := item.(type) {
	case fsVO.FileMetaInfo:
		err := proc.getProc().Process(v)
		if err != nil {
			return err
		}
	default:
		panic(fmt.Sprintf("unrecognized item type in Process: %T", item))
	}

	return nil
}

func (proc *pagesProcessor) Start(ctx context.Context) context.Context {
	return proc.processor.Start(ctx)
}

func (proc *pagesProcessor) Wait() error {
	if err := proc.processor.Wait(); err != nil {
		return err
	}

	return nil
}

type sitePagesProcessor struct {
	m         *PageMap
	ctx       context.Context
	itemChan  chan any
	itemGroup *errgroup.Group

	log log.Logger
}

func (p *sitePagesProcessor) Process(item any) error {
	select {
	case <-p.ctx.Done():
		return nil
	default:
		p.itemChan <- item
	}
	return nil
}

func (p *sitePagesProcessor) Start(ctx context.Context) context.Context {
	p.itemGroup, ctx = errgroup.WithContext(ctx)
	p.ctx = ctx
	p.itemGroup.Go(func() error {
		for item := range p.itemChan {
			if err := p.doProcess(item); err != nil {
				return err
			}
		}
		return nil
	})
	return ctx
}

func (p *sitePagesProcessor) Wait() error {
	close(p.itemChan)
	return p.itemGroup.Wait()
}

func (p *sitePagesProcessor) doProcess(item any) error {
	switch v := item.(type) {
	case fsVO.FileMetaInfo:
		meta := v.Meta()
		classifier := meta.Classifier

		p.log.Printf("doProcess --- %+v\n", meta)

		switch classifier {
		case fsVO.ContentClassContent: //  createOverlayFs
			if err := p.m.AddFilesBundle(v); err != nil {
				return err
			}
		case fsVO.ContentClassFile:
			panic("doProcess not support ContentClassFile yet")
		default:
			panic(fmt.Sprintf("invalid classifier: %q", classifier))
		}
	default:
		panic(fmt.Sprintf("unrecognized item type in Process: %T", item))
	}
	return nil
}

func (proc *pagesProcessor) getProc() pagesCollectorProcessorProvider {
	return proc.processor
}

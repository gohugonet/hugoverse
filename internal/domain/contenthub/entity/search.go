package entity

import (
	"context"
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"sync"
)

type Search struct {
	ctx   context.Context
	pages contenthub.Pages

	relatedDocsHandler *RelatedDocsHandler

	log loggers.Logger
}

func NewSearch(log loggers.Logger) *Search {
	return &Search{
		relatedDocsHandler: NewRelatedDocsHandler(DefaultConfig),
		log:                log,
	}
}

func (p *Search) SearchPage(ctx context.Context, pages contenthub.Pages, page contenthub.Page) (contenthub.Pages, error) {
	return p.SearchPagesWithCtx(ctx, pages, valueobject.SearchOpts{
		Document: page,
	})
}

func (p *Search) SearchPagesWithCtx(ctx context.Context, pages contenthub.Pages, opts valueobject.SearchOpts) (contenthub.Pages, error) {
	p.ctx = ctx
	p.pages = pages
	return p.Search(opts)
}

func (p *Search) Search(opts valueobject.SearchOpts) (contenthub.Pages, error) {
	return p.withInvertedIndex(func(idx *InvertedIndex) ([]contenthub.Document, error) {
		return idx.Search(p.ctx, opts)
	})
}

func (p *Search) withInvertedIndex(search func(idx *InvertedIndex) ([]contenthub.Document, error)) (contenthub.Pages, error) {
	if len(p.pages) == 0 {
		return nil, nil
	}

	searchIndex, err := p.relatedDocsHandler.getOrCreateIndex(p.ctx, p.pages)
	if err != nil {
		p.log.Errorf("error getOrCreateIndex index: %s", err)
		return nil, err
	}

	result, err := search(searchIndex)
	if err != nil {
		p.log.Errorf("error search index: %s", err)
		return nil, err
	}

	if len(result) > 0 {
		mp := make(contenthub.Pages, len(result))
		for i, match := range result {
			mp[i] = match.(contenthub.Page)
		}
		return mp, nil
	}

	return nil, nil
}

type cachedPostingList struct {
	p contenthub.Pages

	postingList *InvertedIndex
}

type RelatedDocsHandler struct {
	cfg Config

	postingLists []*cachedPostingList
	mu           sync.RWMutex
}

func NewRelatedDocsHandler(cfg Config) *RelatedDocsHandler {
	return &RelatedDocsHandler{cfg: cfg}
}

func (s *RelatedDocsHandler) Clone() *RelatedDocsHandler {
	return NewRelatedDocsHandler(s.cfg)
}

// This assumes that a lock has been acquired.
func (s *RelatedDocsHandler) getIndex(p contenthub.Pages) *InvertedIndex {
	for _, ci := range s.postingLists {
		if pagesEqual(p, ci.p) {
			return ci.postingList
		}
	}
	return nil
}

// pagesEqual returns whether p1 and p2 are equal.
func pagesEqual(p1, p2 contenthub.Pages) bool {
	if p1 == nil && p2 == nil {
		return true
	}

	if p1 == nil || p2 == nil {
		return false
	}

	if p1.Len() != p2.Len() {
		return false
	}

	if p1.Len() == 0 {
		return true
	}

	for i := 0; i < len(p1); i++ {
		if p1[i] != p2[i] {
			return false
		}
	}
	return true
}

func (s *RelatedDocsHandler) getOrCreateIndex(ctx context.Context, p contenthub.Pages) (*InvertedIndex, error) {
	s.mu.RLock()
	cachedIndex := s.getIndex(p)
	if cachedIndex != nil {
		s.mu.RUnlock()
		return cachedIndex, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double check.
	if cachedIndex := s.getIndex(p); cachedIndex != nil {
		return cachedIndex, nil
	}

	for _, c := range s.cfg.Indices {
		if c.Type() == valueobject.TypeFragments {
			fmt.Println("TODO: fragments 3")
			break
		}
	}

	searchIndex := NewInvertedIndex(s.cfg)

	for _, page := range p {
		if err := searchIndex.Add(ctx, page); err != nil {
			return nil, err
		}
	}

	s.postingLists = append(s.postingLists, &cachedPostingList{p: p, postingList: searchIndex})

	if err := searchIndex.Finalize(ctx); err != nil {
		return nil, err
	}

	return searchIndex, nil
}

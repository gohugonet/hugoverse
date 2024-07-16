package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"github.com/gohugonet/hugoverse/pkg/output"
)

const (
	StandalonePage404Base     = "404"
	StandalonePageSitemapBase = "_sitemap"
)

type Standalone struct {
	FsSvc contenthub.FsService
	Cache *valueobject.Cache
}

func (s *Standalone) Assemble(pages *doctree.NodeShiftTree[*PageTreesNode], pb *PageBuilder) error {
	key404 := "/" + StandalonePage404Base
	page404, err := s.addStandalone(key404, output.HTTPStatusHTMLFormat, pb)
	if err != nil {
		return err
	}
	pages.InsertIntoValuesDimension(key404, newPageTreesNode(page404))

	keySitemap := "/" + StandalonePageSitemapBase
	pageSitemap, err := s.addStandalone(keySitemap, output.SitemapFormat, pb)
	if err != nil {
		return err
	}
	pages.InsertIntoValuesDimension(keySitemap, newPageTreesNode(pageSitemap))

	return nil
}

func (s *Standalone) addStandalone(key string, format output.Format, pb *PageBuilder) (contenthub.Page, error) {
	fmi := s.FsSvc.NewFileMetaInfo(key + format.MediaType.FirstSuffix.FullSuffix)
	f := valueobject.NewFileInfo(fmi)

	ps, err := newPageSource(f, s.Cache)
	if err != nil {
		return nil, err
	}

	p, err := pb.WithSource(ps).KindBuild()
	if err != nil {
		return nil, err
	}

	return p, nil
}

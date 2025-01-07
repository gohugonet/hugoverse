package entity

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/doctree"
	"strings"
)

const PageHomeBase = "/"

type Section struct {
	home contenthub.Page
	seen map[string]bool

	FsSvc contenthub.FsService
	Cache *Cache
}

func (s *Section) isHome(key string) bool {
	return key == ""
}

func (s *Section) isSectionExist(section string) bool {
	if section == "" || s.seen[section] {
		return true
	}
	s.seen[section] = true
	return false
}

func (s *Section) Assemble(pages *doctree.NodeShiftTree[*PageTreesNode], pb *PageBuilder, langIdx int) error {
	s.seen = make(map[string]bool)

	var w *doctree.NodeShiftTreeWalker[*PageTreesNode]
	w = &doctree.NodeShiftTreeWalker[*PageTreesNode]{
		LockType: doctree.LockTypeWrite,
		Tree:     pages,
		Handle: func(k string, n *PageTreesNode, match doctree.DimensionFlag) (bool, error) {
			if n == nil {
				panic("n is nil")
			}

			ps, found := n.getPage()
			if !found {
				return false, nil
			}

			if s.isHome(k) {
				s.home = ps
				return false, nil
			}

			switch ps.Kind() {
			case valueobject.KindPage, valueobject.KindSection:
				// OK
			default:
				// Skip taxonomy nodes etc.
				return false, nil
			}

			p := ps.Paths()
			section := p.Section()

			if s.isSectionExist(section) {
				return false, nil
			}

			// Try to preserve the original casing if possible.
			sectionUnnormalized := p.Unnormalized().Section()

			fmi := s.FsSvc.NewFileMetaInfo("/" + sectionUnnormalized + "/_index.md")
			f, err := valueobject.NewFileInfo(fmi)
			if err != nil {
				return false, err
			}

			sectionSource, err := newPageSource(f, s.Cache)
			if err != nil {
				return false, err
			}

			sectionBase := sectionSource.Paths().Base()
			nn := w.Tree.Get(sectionBase)

			if nn == nil {
				sectionPage, err := pb.WithSource(sectionSource).WithLangIdx(langIdx).KindBuild()
				if err != nil {
					return false, err
				}

				w.Tree.InsertIntoValuesDimension(sectionBase, newPageTreesNode(sectionPage))
			}

			// /a/b, we don't need to walk deeper.
			if strings.Count(k, "/") > 1 {
				w.SkipPrefix(k + "/")
			}

			return false, nil
		},
	}

	if err := w.Walk(context.Background()); err != nil {
		return err
	}

	if s.home == nil {
		if err := s.CreateHome(pb); err != nil {
			return err
		}

		w.Tree.InsertWithLock(s.home.Paths().Base(), newPageTreesNode(s.home))
	}

	return nil
}

func (s *Section) CreateHome(pb *PageBuilder) error {
	fmi := s.FsSvc.NewFileMetaInfo("/_index.md")
	f, err := valueobject.NewFileInfo(fmi)
	if err != nil {
		return err
	}

	homeSource, err := newPageSource(f, s.Cache)
	if err != nil {
		return err
	}

	homePage, err := pb.WithSource(homeSource).KindBuild()
	if err != nil {
		return err
	}

	s.home = homePage

	return nil
}

package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/lazy"
)

var (
	nopPageOutput = &pageOutput{
		// TODO, simplify
	}
)

func newPageBase(metaProvider *pageMeta) (*pageState, error) {
	ps := &pageState{
		pageOutput: nopPageOutput,
		pageCommon: &pageCommon{
			// Simplify:  FileProvider...
			FileProvider:     metaProvider,
			PageMetaProvider: metaProvider,
			init:             lazy.New(),
			m:                metaProvider,
		},
	}

	return ps, nil
}

func newPage(n *contentNode, kind string, sections ...string) *pageState {
	p, err := newPageFromMeta(
		n,
		&pageMeta{
			kind:     kind,
			sections: sections,
		})
	if err != nil {
		panic(err)
	}

	return p
}

func newPageFromMeta(n *contentNode, metaProvider *pageMeta) (*pageState, error) {
	if metaProvider.f == nil {
		metaProvider.f = valueobject.NewZeroFile()
	}

	ps, err := newPageBase(metaProvider)
	if err != nil {
		return nil, err
	}

	metaProvider.setMetadata()
	metaProvider.applyDefaultValues()

	ps.init.Add(func() (any, error) {
		makeOut := func() *pageOutput {
			return newPageOutput()
		}

		ps.pageOutputs = make([]*pageOutput, 1)
		po := makeOut()
		ps.pageOutputs[0] = po

		contentProvider, err := newPageContentOutput(ps)
		if err != nil {
			return nil, err
		}
		po.initContentProvider(contentProvider)

		return nil, nil
	})

	return ps, err
}

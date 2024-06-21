package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/gohugonet/hugoverse/pkg/maps"
	"github.com/gohugonet/hugoverse/pkg/parser/metadecoders"
	"github.com/gohugonet/hugoverse/pkg/parser/pageparser"
)

type FrontMatter struct {
	Params maps.Params
	*valueobject.Cascade
}

func (fm *FrontMatter) frontMatterMap(it pageparser.Item, source []byte) error {
	f := pageparser.FormatFromFrontMatterType(it.Type)

	m, err := metadecoders.Default.UnmarshalToMap(it.Val(source), f)

	if err != nil {
		return err
	}

	fm.Params = m

	return nil
}

func (fm *FrontMatter) parse() error {
	if err := fm.parseCascade(); err != nil {
		return err
	}
	return nil
}

func (fm *FrontMatter) parseCascade() error {
	if cv, found := fm.Params["cascade"]; found {
		var err error
		fm.Cascade, err = valueobject.NewCascade(cv)
		if err != nil {
			return err
		}
	}
	return nil
}

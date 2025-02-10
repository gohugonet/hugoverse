package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/site/valueobject"
	"github.com/mdfriday/hugoverse/pkg/maps"
)

type Reserve struct {
	site *Site
}

func NewReserve(site *Site) *Reserve {
	return &Reserve{
		site: site,
	}
}

func (r *Reserve) Contact() maps.Params {
	params := maps.Params{}
	lp, err := r.site.GetPage(valueobject.ReservedAboutContactFile)

	if lp != nil && err == nil {
		hs := lp.Result().Headers()
		for _, h := range hs {
			paragraphs := h.Paragraphs()
			if len(paragraphs) > 0 {
				params[h.Name()] = paragraphs[0].Text()
			}
		}
	}

	return params
}

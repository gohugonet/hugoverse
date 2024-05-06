package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/pkg/media"
)

type MediaType struct {
	valueobject.MediaTypeConfig
}

func (t MediaType) LookFirstBySuffix(suffix string) (media.Type, media.SuffixInfo, bool) {
	return t.MediaTypeConfig.Types.GetFirstBySuffix(suffix)
}

func (t MediaType) LookByType(mediaType string) (media.Type, bool) {
	return t.MediaTypeConfig.Types.GetByType(mediaType)
}

package valueobject

import (
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/spf13/afero"
)

type TransformationUpdate struct {
	Content        *string
	SourceFilename *string
	SourceFs       afero.Fs
	TargetPath     string
	MediaType      media.Type
	Data           map[string]any

	StartCtx ResourceTransformationCtx
}

type TransformedResourceMetadata struct {
	Target     string         `json:"Target"`
	MediaTypeV string         `json:"MediaType"`
	MetaData   map[string]any `json:"Data"`
}

func (u *TransformationUpdate) IsContentChanged() bool {
	return u.Content != nil || u.SourceFilename != nil
}

func (u *TransformationUpdate) ToTransformedResourceMetadata() TransformedResourceMetadata {
	return TransformedResourceMetadata{
		MediaTypeV: u.MediaType.Type,
		Target:     u.TargetPath,
		MetaData:   u.Data,
	}
}

func (u *TransformationUpdate) UpdateFromCtx(ctx *ResourceTransformationCtx) {
	u.TargetPath = ctx.OutPath
	u.MediaType = ctx.OutMediaType
	u.Data = ctx.Data
	u.TargetPath = ctx.InPath
}

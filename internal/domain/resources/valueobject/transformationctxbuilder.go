package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"github.com/gohugonet/hugoverse/pkg/media"
)

type ResTransCtxBuilder struct {
	DepSvc resources.DependenceSvc
	PubSvc resources.PublishSvc

	mediaType  media.Type
	src        pio.ReadSeekCloser
	targetPath string
}

func NewResourceTransformationCtxBuilder(d resources.DependenceSvc, p resources.PublishSvc) *ResTransCtxBuilder {
	return &ResTransCtxBuilder{
		DepSvc: d,
		PubSvc: p,
	}
}

func (b *ResTransCtxBuilder) WithMediaType(m media.Type) *ResTransCtxBuilder {
	b.mediaType = m
	return b
}

func (b *ResTransCtxBuilder) WithSource(src pio.ReadSeekCloser) *ResTransCtxBuilder {
	b.src = src
	return b
}

func (b *ResTransCtxBuilder) WithTargetPath(targetPath string) *ResTransCtxBuilder {
	b.targetPath = targetPath
	return b
}

func (b *ResTransCtxBuilder) Build() *ResourceTransformationCtx {
	b1 := bp.GetBuffer()
	b2 := bp.GetBuffer()

	rtc := &ResourceTransformationCtx{
		DepSvc: b.DepSvc,
		PubSvc: b.PubSvc,

		Ctx:  context.Background(),
		Data: make(map[string]any),

		b1: b1,
		b2: b2,

		Source: &TransformationSource{
			From:        b.src,
			InPath:      b.targetPath,
			InMediaType: b.mediaType,
		},
		Target: &TransformationTarget{
			To:           b1,
			OutPath:      "",
			OutMediaType: b.mediaType,
		},
	}

	return rtc
}

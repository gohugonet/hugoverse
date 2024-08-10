package entity

import (
	"context"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/gohugonet/hugoverse/pkg/paths"
)

type TemplateClient struct {
	T Template
}

type executeAsTemplateTransform struct {
	t          Template
	targetPath string
	data       any
}

func (t *executeAsTemplateTransform) Key() valueobject.ResourceTransformationKey {
	return valueobject.NewResourceTransformationKey("execute-as-template", t.targetPath)
}

func (t *executeAsTemplateTransform) Transform(ctx *valueobject.ResourceTransformationCtx) error {
	tplStr := helpers.ReaderToString(ctx.Source.From)
	templ, err := t.t.Parse(ctx.Source.InPath, tplStr)
	if err != nil {
		return fmt.Errorf("failed to parse Resource %q as Template:: %w", ctx.Source.InPath, err)
	}

	ctx.Target.OutPath = t.targetPath

	return t.t.ExecuteWithContext(ctx.Ctx, templ, ctx.Target.To, t.data)
}

func (c *TemplateClient) ExecuteAsTemplate(ctx context.Context, res resources.Resource, targetPath string, data any) (resources.Resource, error) {
	transRes := res.(Transformer)
	return transRes.TransformWithContext(ctx, &executeAsTemplateTransform{
		targetPath: paths.ToSlashTrimLeading(targetPath),
		t:          c.T,
		data:       data,
	})
}

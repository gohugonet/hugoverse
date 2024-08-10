package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/transform"
	"github.com/tdewolff/minify/v2"
)

// MinifierClient for minification of Resource objects. Supported minifiers are:
// css, html, js, json, svg and xml.
type MinifierClient struct {
	M *minify.M

	// Whether output minification is enabled (HTML in /public)
	MinifyOutput bool
}

type minifyTransformation struct {
	m *minify.M
}

func (t *minifyTransformation) Key() valueobject.ResourceTransformationKey {
	return valueobject.NewResourceTransformationKey("minify")
}

func (t *minifyTransformation) Transform(ctx *valueobject.ResourceTransformationCtx) error {
	ctx.AddOutPathIdentifier(".min")
	return t.m.Minify(ctx.Source.InMediaType.Type, ctx.Target.To, ctx.Source.From)
}

func (c *MinifierClient) Minify(res resources.Resource) (resources.Resource, error) {
	transRes := res.(Transformer)
	return transRes.Transform(&minifyTransformation{
		m: c.M,
	})
}

func (c *MinifierClient) Transformer(mediatype media.Type) transform.Transformer {
	_, params, min := c.M.Match(mediatype.Type)
	if min == nil {
		// No minifier for this MIME type
		return nil
	}

	return func(ft transform.FromTo) error {
		// Note that the source io.Reader will already be buffered, but it implements
		// the Bytes() method, which is recognized by the Minify library.
		return min.Minify(c.M, ft.To(), ft.From(), params)
	}
}

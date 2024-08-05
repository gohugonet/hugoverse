package valueobject

import (
	"bytes"
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"io"
	"path"
	"strings"
)

type TransformationSource struct {
	// The Content to transform.
	From io.Reader

	// This is the relative target path to the resource. Unix styled slashes.
	InPath string

	// The input media type
	InMediaType media.Type
}

type TransformationTarget struct {
	// The target of Content transformation.
	// The current implementation requires that r is written to w
	// even if no transformation is performed.
	To io.Writer

	// The relative target path to the transformed resource. Unix styled slashes.
	OutPath string

	// The media type of the transformed resource.
	OutMediaType media.Type
}

type ResourceTransformationCtx struct {
	// The context that started the transformation.
	Ctx context.Context

	DepSvc resources.DependenceSvc
	PubSvc resources.PublishSvc

	Source *TransformationSource
	Target *TransformationTarget

	b1 *bytes.Buffer
	b2 *bytes.Buffer

	// Data data can be set on the transformed Resource. Not that this need
	// to be simple types, as it needs to be serialized to JSON and back.
	Data map[string]any
}

func (ctx *ResourceTransformationCtx) Close() {
	bp.PutBuffer(ctx.b1)
	bp.PutBuffer(ctx.b2)
}

func (ctx *ResourceTransformationCtx) SourcePath() string {
	return strings.TrimPrefix(ctx.Source.InPath, "/")
}

// AddOutPathIdentifier transforming InPath to OutPath adding an identifier,
// eg '.min' before any extension.
func (ctx *ResourceTransformationCtx) AddOutPathIdentifier(identifier string) {
	ctx.Target.OutPath = ctx.addPathIdentifier(ctx.Source.InPath, identifier)
}

// ReplaceOutPathExtension transforming InPath to OutPath replacing the file
// extension, e.g. ".scss"
func (ctx *ResourceTransformationCtx) ReplaceOutPathExtension(newExt string) {
	dir, file := path.Split(ctx.Source.InPath)
	base, _ := paths.PathAndExt(file)
	ctx.Target.OutPath = path.Join(dir, base+newExt)
}

func (ctx *ResourceTransformationCtx) addPathIdentifier(inPath, identifier string) string {
	dir, file := path.Split(inPath)
	base, ext := paths.PathAndExt(file)
	return path.Join(dir, base+identifier+ext)
}

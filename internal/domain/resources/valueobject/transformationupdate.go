package valueobject

import (
	"context"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/media"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"io"
	"path"
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

func (u *TransformationUpdate) isContentChanged() bool {
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

type ResourceTransformationCtx struct {
	// The context that started the transformation.
	Ctx context.Context

	// The dependency manager to use for dependency tracking.
	DependencyManager identity.Manager

	// The Content to transform.
	From io.Reader

	// The target of Content transformation.
	// The current implementation requires that r is written to w
	// even if no transformation is performed.
	To io.Writer

	// This is the relative path to the original source. Unix styled slashes.
	SourcePath string

	// This is the relative target path to the resource. Unix styled slashes.
	InPath string

	// The relative target path to the transformed resource. Unix styled slashes.
	OutPath string

	// The input media type
	InMediaType media.Type

	// The media type of the transformed resource.
	OutMediaType media.Type

	// Data data can be set on the transformed Resource. Not that this need
	// to be simple types, as it needs to be serialized to JSON and back.
	Data map[string]any

	// This is used to publish additional artifacts, e.g. source maps.
	// We may improve this.
	OpenResourcePublisher func(relTargetPath string) (io.WriteCloser, error)
}

// AddOutPathIdentifier transforming InPath to OutPath adding an identifier,
// eg '.min' before any extension.
func (ctx *ResourceTransformationCtx) AddOutPathIdentifier(identifier string) {
	ctx.OutPath = ctx.addPathIdentifier(ctx.InPath, identifier)
}

// PublishSourceMap writes the Content to the target folder of the main resource
// with the ".map" extension added.
func (ctx *ResourceTransformationCtx) PublishSourceMap(content string) error {
	target := ctx.OutPath + ".map"
	f, err := ctx.OpenResourcePublisher(target)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	return err
}

// ReplaceOutPathExtension transforming InPath to OutPath replacing the file
// extension, e.g. ".scss"
func (ctx *ResourceTransformationCtx) ReplaceOutPathExtension(newExt string) {
	dir, file := path.Split(ctx.InPath)
	base, _ := paths.PathAndExt(file)
	ctx.OutPath = path.Join(dir, base+newExt)
}

func (ctx *ResourceTransformationCtx) addPathIdentifier(inPath, identifier string) string {
	dir, file := path.Split(inPath)
	base, ext := paths.PathAndExt(file)
	return path.Join(dir, base+identifier+ext)
}

package resource

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
)

type Resource interface {
	GetResource(pathname string) (resources.Resource, error)
	GetMatch(pattern string) (resources.Resource, error)
	Copy(r resources.Resource, targetPath string) (resources.Resource, error)

	Minify(r resources.Resource) (resources.Resource, error)

	ExecuteAsTemplate(ctx context.Context, res resources.Resource, targetPath string, data any) (resources.Resource, error)

	Fingerprint(res resources.Resource, algo string) (resources.Resource, error)

	ToCSS(res resources.Resource, args map[string]any) (resources.Resource, error)
}

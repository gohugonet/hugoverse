package resource

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
)

type Resource interface {
	GetResource(pathname string) (resources.Resource, error)
	GetMatch(pattern string) (resources.Resource, error)
	Copy(r resources.Resource, targetPath string) (resources.Resource, error)

	Minify(r resources.Resource) (resources.Resource, error)
}

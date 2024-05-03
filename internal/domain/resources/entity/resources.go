package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/hexec"
)

type Resources struct {
	*Creator

	ImageCache *ImageCache

	ExecHelper *hexec.Exec

	*Common
}

func (rs *Resources) GetResource(pathname string) (resources.Resource, error) {
	return nil, nil
}

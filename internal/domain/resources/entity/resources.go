package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/hexec"
)

type Resources struct {
	*Creator

	Imaging    *valueobject.ImageProcessor
	ImageCache *valueobject.ImageCache

	ExecHelper *hexec.Exec

	*Common
}

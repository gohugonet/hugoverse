package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/hexec"
)

type Resources struct {
	*Creator

	ImageCache *valueobject.ImageCache

	ExecHelper *hexec.Exec

	*Common
}

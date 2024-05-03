package valueobject

import (
	"github.com/disintegration/gift"
	"image"
	"image/draw"
)

var _ ImageProcessSpecProvider = (*processFilter)(nil)

type ImageProcessSpecProvider interface {
	ImageProcessSpec() string
}

type processFilter struct {
	spec string
}

func (f processFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	panic("not supported")
}

func (f processFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	panic("not supported")
}

func (f processFilter) ImageProcessSpec() string {
	return f.spec
}

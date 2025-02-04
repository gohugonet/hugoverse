package valueobject

import (
	"github.com/disintegration/gift"
	"image"
	"image/color"
	"image/draw"
)

var _ gift.Filter = (*opacityFilter)(nil)

type opacityFilter struct {
	opacity float32
}

func (f opacityFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	// 0 is fully transparent and 255 is opaque.
	alpha := uint8(f.opacity * 255)
	mask := image.NewUniform(color.Alpha{alpha})
	draw.DrawMask(dst, dst.Bounds(), src, image.Point{}, mask, image.Point{}, draw.Over)
}

func (f opacityFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	return image.Rect(0, 0, srcBounds.Dx(), srcBounds.Dy())
}

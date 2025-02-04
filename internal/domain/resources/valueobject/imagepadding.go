package valueobject

import (
	"github.com/disintegration/gift"
	"image"
	"image/color"
	"image/draw"
)

var _ gift.Filter = (*paddingFilter)(nil)

type paddingFilter struct {
	top, right, bottom, left int
	ccolor                   color.Color // canvas color
}

func (f paddingFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	w := src.Bounds().Dx() + f.left + f.right
	h := src.Bounds().Dy() + f.top + f.bottom

	if w < 1 {
		panic("final image width will be less than 1 pixel: check padding values")
	}
	if h < 1 {
		panic("final image height will be less than 1 pixel: check padding values")
	}

	i := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(i, i.Bounds(), image.NewUniform(f.ccolor), image.Point{}, draw.Src)
	gift.New().Draw(dst, i)
	gift.New().DrawAt(dst, src, image.Pt(f.left, f.top), gift.OverOperator)
}

func (f paddingFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	return image.Rect(0, 0, srcBounds.Dx()+f.left+f.right, srcBounds.Dy()+f.top+f.bottom)
}

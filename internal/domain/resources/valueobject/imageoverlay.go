package valueobject

import (
	"fmt"
	"github.com/disintegration/gift"
	"image"
	"image/draw"
)

var _ gift.Filter = (*overlayFilter)(nil)

type overlayFilter struct {
	src  ImageSource
	x, y int
}

func (f overlayFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	overlaySrc, err := f.src.DecodeImage()
	if err != nil {
		panic(fmt.Sprintf("failed to decode image: %s", err))
	}

	gift.New().Draw(dst, src)
	gift.New().DrawAt(dst, overlaySrc, image.Pt(f.x, f.y), gift.OverOperator)
}

func (f overlayFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	return image.Rect(0, 0, srcBounds.Dx(), srcBounds.Dy())
}

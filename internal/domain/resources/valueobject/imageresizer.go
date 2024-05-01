package valueobject

import (
	"github.com/disintegration/gift"
	"image"
	"math"
)

// Needed by smartcrop
type imagingResizer struct {
	p      *ImageProcessor
	filter gift.Resampling
}

func (r imagingResizer) Resize(img image.Image, width, height uint) image.Image {
	// See https://github.com/gohugoio/hugo/issues/7955#issuecomment-861710681
	scaleX, scaleY := calcFactorsNfnt(width, height, float64(img.Bounds().Dx()), float64(img.Bounds().Dy()))
	if width == 0 {
		width = uint(math.Ceil(float64(img.Bounds().Dx()) / scaleX))
	}
	if height == 0 {
		height = uint(math.Ceil(float64(img.Bounds().Dy()) / scaleY))
	}
	result, _ := r.p.Filter(img, gift.Resize(int(width), int(height), r.filter))
	return result
}

// Calculates scaling factors using old and new image dimensions.
// Code borrowed from https://github.com/nfnt/resize/blob/83c6a9932646f83e3267f353373d47347b6036b2/resize.go#L593
func calcFactorsNfnt(width, height uint, oldWidth, oldHeight float64) (scaleX, scaleY float64) {
	if width == 0 {
		if height == 0 {
			scaleX = 1.0
			scaleY = 1.0
		} else {
			scaleY = oldHeight / float64(height)
			scaleX = scaleY
		}
	} else {
		scaleX = oldWidth / float64(width)
		if height == 0 {
			scaleY = scaleX
		} else {
			scaleY = oldHeight / float64(height)
		}
	}
	return
}

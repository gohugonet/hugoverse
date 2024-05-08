package valueobject

import (
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/gift"
	"image/color"
)

type ImagingConfigInternal struct {
	BgColor        color.Color
	Hint           webpoptions.EncodingPreset
	ResampleFilter gift.Resampling
	Anchor         gift.Anchor

	SourceHash string

	Imaging ImagingConfig
}

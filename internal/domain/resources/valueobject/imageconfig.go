package valueobject

import (
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/gift"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"image/color"
	"strconv"
)

const (
	// Do not change.
	smartCropIdentifier = "smart"

	// This is just a increment, starting on 1. If Smart Crop improves its cropping, we
	// need a way to trigger a re-generation of the crops in the wild, so increment this.
	smartCropVersionNumber = 1
)

const (
	ActionResize = "resize"
	ActionCrop   = "crop"
	ActionFit    = "fit"
	ActionFill   = "fill"
)

var (
	// Add or increment if changes to an image format's processing requires
	// re-generation.
	imageFormatsVersions = map[resources.ImageFormat]int{
		resources.PNG:  3, // Fix transparency issue with 32 bit images.
		resources.WEBP: 2, // Fix transparency issue with 32 bit images.
		resources.GIF:  1, // Fix resize issue with animated GIFs when target != GIF.
	}

	mainImageVersionNumber = 0
)

// ImageConfig holds configuration to create a new image from an existing one, resize etc.
type ImageConfig struct {
	// This defines the output format of the output image. It defaults to the source format.
	TargetFormat resources.ImageFormat

	Action string

	// If set, this will be used as the key in filenames etc.
	Key string

	// Quality ranges from 1 to 100 inclusive, higher is better.
	// This is only relevant for JPEG and WEBP images.
	// Default is 75.
	Quality            int
	qualitySetForImage bool // Whether the above is set for this image.

	// Rotate rotates an image by the given angle counter-clockwise.
	// The rotation will be performed first.
	Rotate int

	// Used to fill any transparency.
	// When set in site config, it's used when converting to a format that does
	// not support transparency.
	// When set per image operation, it's used even for formats that does support
	// transparency.
	BgColor    color.Color
	BgColorStr string

	// Hint about what type of picture this is. Used to optimize encoding
	// when target is set to webp.
	Hint webpoptions.EncodingPreset

	Width  int
	Height int

	Filter    gift.Resampling
	FilterStr string

	Anchor    gift.Anchor
	AnchorStr string
}

func (i ImageConfig) GetKey(format resources.ImageFormat) string {
	if i.Key != "" {
		return i.Action + "_" + i.Key
	}

	k := strconv.Itoa(i.Width) + "x" + strconv.Itoa(i.Height)
	if i.Action != "" {
		k += "_" + i.Action
	}
	// This slightly odd construct is here to preserve the old image keys.
	if i.qualitySetForImage || RequiresDefaultQuality(i.TargetFormat) {
		k += "_q" + strconv.Itoa(i.Quality)
	}
	if i.Rotate != 0 {
		k += "_r" + strconv.Itoa(i.Rotate)
	}
	if i.BgColorStr != "" {
		k += "_bg" + i.BgColorStr
	}

	if i.TargetFormat == resources.WEBP {
		k += "_h" + strconv.Itoa(int(i.Hint))
	}

	anchor := i.AnchorStr
	if anchor == smartCropIdentifier {
		anchor = anchor + strconv.Itoa(smartCropVersionNumber)
	}

	k += "_" + i.FilterStr

	if i.Action == ActionFill || i.Action == ActionCrop {
		k += "_" + anchor
	}

	if v, ok := imageFormatsVersions[format]; ok {
		k += "_" + strconv.Itoa(v)
	}

	if mainImageVersionNumber > 0 {
		k += "_" + strconv.Itoa(mainImageVersionNumber)
	}

	return k
}

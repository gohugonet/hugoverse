package valueobject

import (
	"errors"
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/gift"
	"github.com/mdfriday/hugoverse/internal/domain/resources"
	"github.com/mdfriday/hugoverse/pkg/images"
	"image/color"
	"strconv"
	"strings"
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
	imageFormats = map[string]resources.ImageFormat{
		".jpg":  resources.JPEG,
		".jpeg": resources.JPEG,
		".jpe":  resources.JPEG,
		".jif":  resources.JPEG,
		".jfif": resources.JPEG,
		".png":  resources.PNG,
		".tif":  resources.TIFF,
		".tiff": resources.TIFF,
		".bmp":  resources.BMP,
		".gif":  resources.GIF,
		".webp": resources.WEBP,
	}

	// Add or increment if changes to an images format's processing requires
	// re-generation.
	imageFormatsVersions = map[resources.ImageFormat]int{
		resources.PNG:  3, // Fix transparency issue with 32 bit images.
		resources.WEBP: 2, // Fix transparency issue with 32 bit images.
		resources.GIF:  1, // Fix resize issue with animated GIFs when target != GIF.
	}

	mainImageVersionNumber = 0
)

func ImageFormatFromExt(ext string) (resources.ImageFormat, bool) {
	f, found := imageFormats[ext]
	return f, found
}

// ImageConfig holds configuration to create a new images from an existing one, resize etc.
type ImageConfig struct {
	// This defines the output format of the output images. It defaults to the source format.
	TargetFormat resources.ImageFormat

	Action string

	// If set, this will be used as the key in filenames etc.
	Key string

	// Quality ranges from 1 to 100 inclusive, higher is better.
	// This is only relevant for JPEG and WEBP images.
	// Default is 75.
	Quality            int
	qualitySetForImage bool // Whether the above is set for this images.

	// Rotate rotates an images by the given angle counter-clockwise.
	// The rotation will be performed first.
	Rotate int

	// Used to fill any transparency.
	// When set in site config, it's used when converting to a format that does
	// not support transparency.
	// When set per images operation, it's used even for formats that does support
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
	// This slightly odd construct is here to preserve the old images keys.
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

func DecodeImageConfig(action string, options []string, defaults resources.ImageConfig, sourceFormat resources.ImageFormat) (ImageConfig, error) {
	var (
		c   ImageConfig = GetDefaultImageConfig(action, defaults)
		err error
	)

	action = strings.ToLower(action)

	c.Action = action

	if options == nil {
		return c, errors.New("image options cannot be empty")
	}

	for _, part := range options {
		part = strings.ToLower(part)

		if part == smartCropIdentifier {
			c.AnchorStr = smartCropIdentifier
		} else if pos, ok := images.AnchorPositions[part]; ok {
			c.Anchor = pos
			c.AnchorStr = part
		} else if filter, ok := images.ImageFilters[part]; ok {
			c.Filter = filter
			c.FilterStr = part
		} else if hint, ok := images.Hints[part]; ok {
			c.Hint = hint
		} else if part[0] == '#' {
			c.BgColorStr = part[1:]
			c.BgColor, err = images.HexStringToColor(c.BgColorStr)
			if err != nil {
				return c, err
			}
		} else if part[0] == 'q' {
			c.Quality, err = strconv.Atoi(part[1:])
			if err != nil {
				return c, err
			}
			if c.Quality < 1 || c.Quality > 100 {
				return c, errors.New("quality ranges from 1 to 100 inclusive")
			}
			c.qualitySetForImage = true
		} else if part[0] == 'r' {
			c.Rotate, err = strconv.Atoi(part[1:])
			if err != nil {
				return c, err
			}
		} else if strings.Contains(part, "x") {
			widthHeight := strings.Split(part, "x")
			if len(widthHeight) <= 2 {
				first := widthHeight[0]
				if first != "" {
					c.Width, err = strconv.Atoi(first)
					if err != nil {
						return c, err
					}
				}

				if len(widthHeight) == 2 {
					second := widthHeight[1]
					if second != "" {
						c.Height, err = strconv.Atoi(second)
						if err != nil {
							return c, err
						}
					}
				}
			} else {
				return c, errors.New("invalid image dimensions")
			}
		} else if f, ok := ImageFormatFromExt("." + part); ok {
			c.TargetFormat = f
		}
	}

	switch c.Action {
	case ActionCrop, ActionFill, ActionFit:
		if c.Width == 0 || c.Height == 0 {
			return c, errors.New("must provide Width and Height")
		}
	case ActionResize:
		if c.Width == 0 && c.Height == 0 {
			return c, errors.New("must provide Width or Height")
		}
	default:
		if c.Width != 0 || c.Height != 0 {
			return c, errors.New("width or height are not supported for this action")
		}
	}

	if action != "" && c.FilterStr == "" {
		c.FilterStr = defaults.ResamplingStr()
		c.Filter = defaults.Resampling()
	}

	if c.Hint == 0 {
		c.Hint = webpoptions.EncodingPresetPhoto
	}

	if action != "" && c.AnchorStr == "" {
		c.AnchorStr = defaults.AnchorStr()
		c.Anchor = defaults.Anchor()
	}

	// default to the source format
	if c.TargetFormat == 0 {
		c.TargetFormat = sourceFormat
	}

	if c.Quality <= 0 && RequiresDefaultQuality(c.TargetFormat) {
		// We need a quality setting for all JPEGs and WEBPs.
		c.Quality = defaults.ImageQuality()
	}

	if c.BgColor == nil && c.TargetFormat != sourceFormat {
		if SupportsTransparency(sourceFormat) && !SupportsTransparency(c.TargetFormat) {
			c.BgColor = defaults.BgColor()
			c.BgColorStr = defaults.BgColorStr()
		}
	}

	return c, nil
}

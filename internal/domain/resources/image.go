package resources

import (
	"github.com/mdfriday/hugoverse/pkg/images/exif"
	"github.com/mdfriday/hugoverse/pkg/media"
	"image"
	"image/gif"
)

// ImageFormat is an images file format.
type ImageFormat int

const (
	JPEG ImageFormat = iota + 1
	PNG
	GIF
	TIFF
	BMP
	WEBP
)

var (
	ImageFormatsBySubType = map[string]ImageFormat{
		media.Builtin.JPEGType.Sub(): JPEG,
		media.Builtin.PNGType.Sub():  PNG,
		media.Builtin.TIFFType.Sub(): TIFF,
		media.Builtin.BMPType.Sub():  BMP,
		media.Builtin.GIFType.Sub():  GIF,
		media.Builtin.WEBPType.Sub(): WEBP,
	}
)

// Giphy represents a GIF Image that may be animated.
type Giphy interface {
	image.Image    // The first frame.
	GIF() *gif.GIF // All frames.
}

type ImageResourceOps interface {
	// Height returns the height of the Image.
	Height() int

	// Width returns the width of the Image.
	Width() int

	// Process applies the given images processing options to the images.
	Process(spec string) (ImageResource, error)

	// Crop an images to match the given dimensions without resizing.
	// You must provide both width and height.
	// Use the anchor option to change the crop box anchor point.
	//    {{ $images := $images.Crop "600x400" }}
	Crop(spec string) (ImageResource, error)

	// Fill scales the images to the smallest possible size that will cover the specified dimensions in spec,
	// crops the resized images to the specified dimensions using the given anchor point.
	// The spec is space delimited, e.g. `200x300 TopLeft`.
	Fill(spec string) (ImageResource, error)

	// Fit scales down the images using the given spec.
	Fit(spec string) (ImageResource, error)

	// Resize resizes the images to the given spec. If one of width or height is 0, the images aspect
	// ratio is preserved.
	Resize(spec string) (ImageResource, error)

	// Filter applies one or more filters to an Image.
	//    {{ $images := $images.Filter (images.GaussianBlur 6) (images.Pixelate 8) }}
	Filter(filters ...any) (ImageResource, error)

	// Exif returns an ExifInfo object containing Image metadata.
	Exif() *exif.ExifInfo

	// Colors returns a slice of the most dominant colors in an images
	// using a simple histogram method.
	Colors() ([]string, error)

	// For internal use.
	DecodeImage() (image.Image, error)
}

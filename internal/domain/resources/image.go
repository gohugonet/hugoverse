package resources

import (
	"github.com/gohugonet/hugoverse/pkg/media"
	"image"
	"image/gif"
)

// ImageFormat is an image file format.
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
		media.Builtin.JPEGType.SubType: JPEG,
		media.Builtin.PNGType.SubType:  PNG,
		media.Builtin.TIFFType.SubType: TIFF,
		media.Builtin.BMPType.SubType:  BMP,
		media.Builtin.GIFType.SubType:  GIF,
		media.Builtin.WEBPType.SubType: WEBP,
	}
)

// Giphy represents a GIF Image that may be animated.
type Giphy interface {
	image.Image    // The first frame.
	GIF() *gif.GIF // All frames.
}

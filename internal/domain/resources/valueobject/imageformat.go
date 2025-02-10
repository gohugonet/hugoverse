package valueobject

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/resources"
	"github.com/mdfriday/hugoverse/pkg/media"
)

// RequiresDefaultQuality returns if the default quality needs to be applied to
// images of this format.
func RequiresDefaultQuality(f resources.ImageFormat) bool {
	return f == resources.JPEG || f == resources.WEBP
}

// SupportsTransparency reports whether it supports transparency in any form.
func SupportsTransparency(f resources.ImageFormat) bool {
	return f != resources.JPEG
}

// DefaultExtension returns the default file extension of this format, starting with a dot.
// For example: .jpg for resources.JPEG
func DefaultExtension(f resources.ImageFormat) string {
	return MediaType(f).FirstSuffix.FullSuffix
}

// MediaType returns the media type of this images, e.g. images/jpeg for resources.JPEG
func MediaType(f resources.ImageFormat) media.Type {
	switch f {
	case resources.JPEG:
		return media.Builtin.JPEGType
	case resources.PNG:
		return media.Builtin.PNGType
	case resources.GIF:
		return media.Builtin.GIFType
	case resources.TIFF:
		return media.Builtin.TIFFType
	case resources.BMP:
		return media.Builtin.BMPType
	case resources.WEBP:
		return media.Builtin.WEBPType
	default:
		panic(fmt.Sprintf("%d is not a valid images format", f))
	}
}

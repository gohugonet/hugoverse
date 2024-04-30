package entity

import "github.com/gohugonet/hugoverse/pkg/image/exif"

type ImageProcessor struct {
	ExifDecoder *exif.Decoder
}

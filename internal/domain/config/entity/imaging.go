package entity

import (
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"github.com/gohugonet/hugoverse/pkg/image/exif"
)

type Imaging struct {
	valueobject.ImagingConfigInternal
}

func (i Imaging) ExifDecoder() (*exif.Decoder, error) {
	exifDecoder, err := exif.NewDecoder(
		exif.WithDateDisabled(i.ImagingConfigInternal.Imaging.Exif.DisableDate),
		exif.WithLatLongDisabled(i.ImagingConfigInternal.Imaging.Exif.DisableLatLong),
		exif.ExcludeFields(i.ImagingConfigInternal.Imaging.Exif.ExcludeFields),
		exif.IncludeFields(i.ImagingConfigInternal.Imaging.Exif.IncludeFields),
	)
	if err != nil {
		return nil, err
	}
	return exifDecoder, nil
}

func (i Imaging) ImageHint() webpoptions.EncodingPreset {
	return i.Hint
}

func (i Imaging) ImageQuality() int {
	return i.Imaging.Quality
}
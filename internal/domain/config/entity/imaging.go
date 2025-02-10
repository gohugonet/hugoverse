package entity

import (
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/gift"
	"github.com/mdfriday/hugoverse/internal/domain/config/valueobject"
	"github.com/mdfriday/hugoverse/pkg/images/exif"
	"image/color"
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

func (i Imaging) ImageHint() webpoptions.EncodingPreset { return i.Hint }
func (i Imaging) ImageQuality() int                     { return i.Imaging.Quality }
func (i Imaging) Resampling() gift.Resampling           { return i.ResampleFilter }
func (i Imaging) ResamplingStr() string                 { return i.Imaging.ResampleFilter }
func (i Imaging) Anchor() gift.Anchor                   { return i.ImagingConfigInternal.Anchor }
func (i Imaging) AnchorStr() string                     { return i.Imaging.Anchor }
func (i Imaging) BgColor() color.Color                  { return i.ImagingConfigInternal.BgColor }
func (i Imaging) BgColorStr() string                    { return i.Imaging.BgColor }
func (i Imaging) SourceHash() string                    { return i.ImagingConfigInternal.SourceHash }

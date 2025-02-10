package valueobject

import (
	"github.com/disintegration/gift"
	"github.com/mdfriday/hugoverse/pkg/images/exif"
	"image"
	"image/draw"
)

type ImageFilterFromOrientationProvider interface {
	AutoOrient(exifInfo *exif.ExifInfo) gift.Filter
}

var _ gift.Filter = (*autoOrientFilter)(nil)

var transformationFilters = map[int]gift.Filter{
	2: gift.FlipHorizontal(),
	3: gift.Rotate180(),
	4: gift.FlipVertical(),
	5: gift.Transpose(),
	6: gift.Rotate270(),
	7: gift.Transverse(),
	8: gift.Rotate90(),
}

type autoOrientFilter struct{}

func (f autoOrientFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	panic("not supported")
}

func (f autoOrientFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	panic("not supported")
}

func (f autoOrientFilter) AutoOrient(exifInfo *exif.ExifInfo) gift.Filter {
	if exifInfo != nil {
		if orientation, ok := exifInfo.Tags["Orientation"].(int); ok {
			if filter, ok := transformationFilters[orientation]; ok {
				return filter
			}
		}
	}

	return nil
}

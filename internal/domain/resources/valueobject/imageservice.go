package valueobject

import "github.com/gohugonet/hugoverse/internal/domain/resources"

func ImageFormatFromMediaSubType(sub string) (resources.ImageFormat, bool) {
	f, found := resources.ImageFormatsBySubType[sub]
	return f, found
}

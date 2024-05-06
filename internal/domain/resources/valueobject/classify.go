package valueobject

import "github.com/gohugonet/hugoverse/pkg/media"

func ClassifyType(t string) string {
	for _, it := range media.BuiltinImages {
		if t == it.Type {
			return "image"
		}
	}

	for _, ct := range media.BuiltinCss {
		if t == ct.Type {
			return "transformer"
		}
	}

	return "general"
}

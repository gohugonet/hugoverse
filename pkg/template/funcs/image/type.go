package image

import "github.com/disintegration/gift"

type Image interface {
	AutoOrient() gift.Filter
}

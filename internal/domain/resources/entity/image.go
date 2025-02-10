package entity

import "github.com/mdfriday/hugoverse/internal/domain/resources/valueobject"

type Image struct {
	*valueobject.Filters
}

func NewImage() *Image {
	return &Image{
		Filters: &valueobject.Filters{},
	}
}

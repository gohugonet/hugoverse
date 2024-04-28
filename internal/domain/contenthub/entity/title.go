package entity

import (
	"github.com/jdkato/prose/transform"
)

type TitleStyle string

const (
	StyleNone TitleStyle = ""
	StyleAP              = "ap"
)

type Title struct {
	Style TitleStyle
}

func (t *Title) CreateTitle(raw string) string {
	switch t.Style {
	case StyleAP:
		tc := transform.NewTitleConverter(transform.APStyle)
		return tc.Title(raw)
	default:
		return raw
	}
}

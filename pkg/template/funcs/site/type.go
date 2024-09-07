package site

import "github.com/gohugonet/hugoverse/pkg/maps"

type Service interface {
	Author
	Meta
	GoogleAnalytics
}

type Author interface {
	Name() string
	Email() string
}

type Meta interface {
	Params() maps.Params
}

type GoogleAnalytics interface {
	IsGoogleAnalyticsEnabled() bool
	GoogleAnalyticsID() string
	IsGoogleAnalyticsRespectDoNotTrack() bool
}

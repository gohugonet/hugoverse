package site

import "github.com/mdfriday/hugoverse/pkg/maps"

type Service interface {
	Author
	Meta
	GoogleAnalytics
	Disqus
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

type Disqus interface {
	IsDisqusEnabled() bool
	DisqusShortname() string
}

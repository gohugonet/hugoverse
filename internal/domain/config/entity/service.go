package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
)

type Service struct {
	valueobject.ServiceConfig
}

func (s Service) IsGoogleAnalyticsEnabled() bool {
	return !s.GoogleAnalytics.Disable
}

func (s Service) GoogleAnalyticsID() string {
	return s.GoogleAnalytics.ID
}

func (s Service) IsGoogleAnalyticsRespectDoNotTrack() bool {
	return s.GoogleAnalytics.RespectDoNotTrack
}

func (s Service) IsDisqusEnabled() bool {
	return !s.Disqus.Disable
}

func (s Service) DisqusShortname() string {
	return s.Disqus.Shortname
}

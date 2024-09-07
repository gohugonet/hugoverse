package entity

import "github.com/gohugonet/hugoverse/internal/domain/config/valueobject"

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

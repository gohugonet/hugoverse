package cache

import (
	"fmt"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"net/http"
	"strings"
)

type Controller interface {
	CacheDisabled() bool
	CacheMaxAge() int64
	ETage() string
}

type Cache struct {
	adminApp Controller
	log      loggers.Logger
}

func New(log loggers.Logger, adminApp Controller) *Cache {
	return &Cache{
		adminApp: adminApp,
		log:      log,
	}
}

const (
	// DefaultMaxAge provides a 2592000 second (30-day) cache max-age setting
	DefaultMaxAge = int64(60 * 60 * 24 * 30)
)

// Control sets the default cache policy on static asset responses
func (s *Cache) Control(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if s.adminApp.CacheDisabled() {
			res.Header().Add("Cache-Control", "no-cache")
			next.ServeHTTP(res, req)
		} else {
			age := s.adminApp.CacheMaxAge()
			etag := s.adminApp.ETage()
			if age == 0 {
				age = DefaultMaxAge
			}
			policy := fmt.Sprintf("max-age=%d, public", age)
			res.Header().Add("ETag", etag)
			res.Header().Add("Cache-Control", policy)

			if match := req.Header.Get("If-None-Match"); match != "" {
				if strings.Contains(match, etag) {
					res.WriteHeader(http.StatusNotModified)
					return
				}
			}

			next.ServeHTTP(res, req)
		}
	})
}

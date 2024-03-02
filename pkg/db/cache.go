package db

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	// DefaultMaxAge provides a 2592000 second (30-day) cache max-age setting
	DefaultMaxAge = int64(60 * 60 * 24 * 30)
)

// CacheControl sets the default cache policy on static asset responses
func CacheControl(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cacheDisabled := ConfigCache("cache_disabled").(bool)
		if cacheDisabled {
			res.Header().Add("Cache-Control", "no-cache")
			next.ServeHTTP(res, req)
		} else {
			age := int64(ConfigCache("cache_max_age").(float64))
			etag := ConfigCache("etag").(string)
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

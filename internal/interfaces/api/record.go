package api

import (
	"github.com/gohugonet/hugoverse/internal/interfaces/api/analytics"
	"net/http"
)

// Record wraps a HandlerFunc to record API requests for analytical purposes
func Record(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		go analytics.Record(req)

		next.ServeHTTP(res, req)
	})
}

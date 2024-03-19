package record

import (
	"github.com/gohugonet/hugoverse/internal/interfaces/api/record/analytics"
	"net/http"
)

type Record struct {
	dataDir string
}

func New(dataDir string) *Record {
	return &Record{
		dataDir: dataDir,
	}
}

func (r *Record) Start() {
	analytics.Setup(r.dataDir)
}

func (r *Record) Close() {
	analytics.Close()
}

// Collect wraps a HandlerFunc to record API requests for analytical purposes
func (r *Record) Collect(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		go analytics.Record(req)

		next.ServeHTTP(res, req)
	})
}

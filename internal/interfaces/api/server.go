package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/application"
	"github.com/gohugonet/hugoverse/pkg/db"
	"github.com/gohugonet/hugoverse/pkg/log"
	"io"
	"net/http"
)

type Server struct {
	mux   *http.ServeMux
	cache responseCache
	Log   log.Logger
}

func NewServer(options ...func(s *Server) error) (*Server, error) {
	s := &Server{mux: http.NewServeMux()}
	for _, o := range options {
		if err := o(s); err != nil {
			return nil, err
		}
	}
	if s.Log == nil {
		return nil, fmt.Errorf("must provide an option func that specifies a logger")
	}
	s.registerHandler()

	csApp := application.NewContentServer()
	db.Start(csApp.DataDir(), csApp.AllContentTypeNames())

	return s, nil
}

func (s *Server) registerHandler() {
	s.mux.HandleFunc("/demo", s.handleDemo)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Forwarded-Proto") == "http" {
		r.URL.Scheme = "https"
		r.URL.Host = r.Host
		http.Redirect(w, r, r.URL.String(), http.StatusFound)
		return
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; preload")
	}
	s.mux.ServeHTTP(w, r)
}

// writeJSONResponse JSON-encodes resp and writes to w with the given HTTP
// status.
func (s *Server) writeJSONResponse(w http.ResponseWriter, resp interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(resp); err != nil {
		s.Log.Errorf("error encoding response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	if _, err := io.Copy(w, &buf); err != nil {
		s.Log.Errorf("io.Copy(w, &buf): %v", err)
		return
	}
}

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/application"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/admin"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/analytics"
	"github.com/gohugonet/hugoverse/pkg/log"
	"io"
	"net/http"
)

type PORT string

const (
	HttpsPort PORT = "https_port"
	HttpPort  PORT = "http_port"
)

const BindAddress = "bind_addr"

type ENV int

const (
	DEV ENV = iota
	PROD
)

type Server struct {
	mux        *http.ServeMux
	cache      responseCache
	Log        log.Logger
	Bind       string
	HttpsPort  int
	HttpPort   int
	db         *database
	contentApp *application.ContentServer
	adminApp   *application.AdminServer
	adminView  *admin.View
}

func (s *Server) ListenAndServe(env ENV, enableHttps bool) error {
	if enableHttps {
		if err := s.enableTLS(env); err != nil {
			s.Log.Fatalf("System failed to enable TLS. Please try to run again.", err)
			return err
		}
	}
	if err := s.saveConfig(); err != nil {
		s.Log.Fatalf("System failed to save config. Please try to run again.", err)
		return err
	}

	s.Log.Printf("Listening on %s:%d", s.Bind, s.HttpPort)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.Bind, s.HttpPort), s)
}

func NewServer(options ...func(s *Server) error) (*Server, error) {
	s := &Server{
		mux:       http.NewServeMux(),
		db:        &database{},
		Bind:      "localhost",
		HttpPort:  80,
		HttpsPort: 443,
	}
	for _, o := range options {
		if err := o(s); err != nil {
			return nil, err
		}
	}
	if s.Log == nil {
		return nil, fmt.Errorf("must provide an option func that specifies a logger")
	}
	s.registerHandler()

	s.contentApp = application.NewContentServer(s.db)

	s.db.start(s.contentApp.AllContentTypeNames())

	server, err := application.NewAdminServer(s.db)
	if err != nil {
		return nil, err
	}
	s.adminApp = server
	s.adminView = admin.NewView(s.adminApp.Name(), s.contentApp.AllContentTypes())

	analytics.Setup(dataDir())

	return s, nil
}

func (s *Server) Close() {
	s.db.close()
	analytics.Close()
}

func (s *Server) registerHandler() {
	s.mux.HandleFunc("/api/demo", s.handleDemo)

	s.registerContentHandler()
	s.registerAdminHandler()
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

func (s *Server) enableTLS(env ENV) error {
	switch env {
	case DEV:
		// todo
		return nil
	case PROD:
		// todo
		return nil
	}
	return nil
}

func (s *Server) saveConfig() error {
	err := s.adminApp.PutConfig(string(HttpsPort), s.HttpsPort)
	if err != nil {
		s.Log.Fatalf("System failed to save Https Port config. Please try to run again.", err)
		return err
	}
	err = s.adminApp.PutConfig(string(HttpPort), s.HttpPort)
	if err != nil {
		s.Log.Fatalf("System failed to save Http Port config. Please try to run again.", err)
		return err
	}
	err = s.adminApp.PutConfig(string(BindAddress), s.Bind)
	if err != nil {
		s.Log.Fatalf("System failed to save bind address config. Please try to run again.", err)
		return err
	}
	return nil
}

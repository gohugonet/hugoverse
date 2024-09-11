package static

import (
	"context"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/spf13/afero"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type FileServer struct {
	PublishDir afero.Fs

	server       *Server
	portListener *serverPortListener

	log loggers.Logger
}

func NewFileServer(dest afero.Fs) *FileServer {
	return &FileServer{
		PublishDir:   dest,
		server:       newServer(),
		portListener: newDefaultServerPortListener(),
		log:          loggers.NewDefault(),
	}
}

func (s *FileServer) Serve() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	mu, err := s.createEndpoint()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    s.portListener.endpoint,
		Handler: mu,
	}

	wg1, ctx := errgroup.WithContext(context.Background())
	wg1.Go(func() error {
		err = srv.Serve(s.portListener.ln)
		if err != nil && !errors.Is(http.ErrServerClosed, err) {
			return err
		}
		return nil
	})

	s.log.Println("Press Ctrl+C to stop")

	err = func() error {
		for {
			select {
			case <-sigs:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}()
	if err != nil {
		s.log.Errorln(err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wg2, ctx := errgroup.WithContext(ctx)
	wg2.Go(func() error {
		return srv.Shutdown(ctx)
	})

	err1, err2 := wg1.Wait(), wg2.Wait()
	if err1 != nil {
		return err1
	}
	return err2
}

func (s *FileServer) createEndpoint() (*http.ServeMux, error) {
	httpFs := afero.NewHttpFs(s.PublishDir)
	fs := filesOnlyFs{httpFs.Dir("/")}
	handler := s.decorateHandler(http.FileServer(fs))

	mu := http.NewServeMux()
	mu.Handle("/", handler)

	return mu, nil
}

func (s *FileServer) decorateHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")

		// Ignore any query params for the operations below.
		requestURI, _ := url.PathUnescape(strings.TrimSuffix(r.RequestURI, "?"+r.URL.RawQuery))
		for _, header := range s.server.MatchHeaders(requestURI) {
			w.Header().Set(header.Key, header.Value)
		}

		if redirect := s.server.MatchRedirect(requestURI); !redirect.IsZero() {
			// fullName := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))
			doRedirect := true
			// This matches Netlify's behaviour and is needed for SPA behaviour.
			// See https://docs.netlify.com/routing/redirects/rewrites-proxies/
			if !redirect.Force {
				path := filepath.Clean(strings.TrimPrefix(requestURI, s.portListener.bathPath()))
				fi, err := s.PublishDir.Stat(path)

				if err == nil {
					if fi.IsDir() {
						// There will be overlapping directories, so we
						// need to check for a file.
						_, err = s.PublishDir.Stat(filepath.Join(path, "index.html"))
						doRedirect = err != nil
					} else {
						doRedirect = false
					}
				}
			}

			if doRedirect {
				switch redirect.Status {
				case 404:
					w.WriteHeader(404)
					file, err := s.PublishDir.Open(strings.TrimPrefix(redirect.To, s.portListener.bathPath()))
					if err == nil {
						defer file.Close()
						io.Copy(w, file)
					} else {
						fmt.Fprintln(w, "<h1>Page Not Found</h1>")
					}
					return
				case 200:
					if r2 := s.rewriteRequest(r, strings.TrimPrefix(redirect.To, s.portListener.bathPath())); r2 != nil {
						requestURI = redirect.To
						r = r2
					}
				default:
					w.Header().Set("Content-Type", "")
					http.Redirect(w, r, redirect.To, redirect.Status)
					return

				}
			}

		}

		h.ServeHTTP(w, r)
	})
}

func (s *FileServer) rewriteRequest(r *http.Request, toPath string) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.URL = new(url.URL)
	*r2.URL = *r.URL
	r2.URL.Path = toPath
	r2.Header.Set("X-Rewrite-Original-URI", r.URL.RequestURI())

	return r2
}

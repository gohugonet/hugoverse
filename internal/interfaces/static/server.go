package static

import (
	"github.com/gobwas/glob"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/mdfriday/hugoverse/pkg/types"
	"github.com/spf13/cast"
	"sort"
	"strings"
)

type Server struct {
	Headers   []Headers
	Redirects []Redirect

	compiledHeaders   []glob.Glob
	compiledRedirects []glob.Glob
}

func newServer() *Server {
	return &Server{
		Redirects: []Redirect{
			{
				From:   "**",
				To:     "/404.html",
				Status: 404,
			},
		},
	}
}

func (s *Server) CompileConfig(logger loggers.Logger) error {
	if s.compiledHeaders != nil {
		return nil
	}
	for _, h := range s.Headers {
		s.compiledHeaders = append(s.compiledHeaders, glob.MustCompile(h.For))
	}
	for _, r := range s.Redirects {
		s.compiledRedirects = append(s.compiledRedirects, glob.MustCompile(r.From))
	}

	return nil
}

func (s *Server) MatchHeaders(pattern string) []types.KeyValueStr {
	if s.compiledHeaders == nil {
		return nil
	}

	var matches []types.KeyValueStr

	for i, g := range s.compiledHeaders {
		if g.Match(pattern) {
			h := s.Headers[i]
			for k, v := range h.Values {
				matches = append(matches, types.KeyValueStr{Key: k, Value: cast.ToString(v)})
			}
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Key < matches[j].Key
	})

	return matches
}

func (s *Server) MatchRedirect(pattern string) Redirect {
	if s.compiledRedirects == nil {
		return Redirect{}
	}

	pattern = strings.TrimSuffix(pattern, "index.html")

	for i, g := range s.compiledRedirects {
		redir := s.Redirects[i]

		// No redirect to self.
		if redir.To == pattern {
			return Redirect{}
		}

		if g.Match(pattern) {
			return redir
		}
	}

	return Redirect{}
}

type Headers struct {
	For    string
	Values map[string]any
}

type Redirect struct {
	From string
	To   string

	// HTTP status code to use for the redirect.
	// A status code of 200 will trigger a URL rewrite.
	Status int

	// Forcode redirect, even if original request path exists.
	Force bool
}

func (r Redirect) IsZero() bool {
	return r.From == ""
}

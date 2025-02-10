package tls

import (
	"github.com/mdfriday/hugoverse/internal/domain/admin"
	"net/http"
)

type Tls struct {
	handler http.Handler
	http    admin.Http
	dir     string
}

func NewTls(handler http.Handler, http admin.Http, dir string) *Tls {
	return &Tls{
		handler: handler,
		http:    http,
		dir:     dir,
	}
}

func (tls *Tls) EnableDev() error {
	return tls.enableDev()
}

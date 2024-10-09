package auth

import (
	"github.com/gohugonet/hugoverse/internal/interfaces/api/token"
	"net/http"
)

type Auth struct {
}

// Check is HTTP middleware to ensure the request has proper token credentials
func (a *Auth) Check(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if a.IsValid(req) {
			next.ServeHTTP(res, req)
			return
		}

		res.WriteHeader(http.StatusUnauthorized)
	})
}

func (a *Auth) CheckWithRedirect(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		redirect := req.URL.Scheme + req.URL.Host + "/admin/login"

		if a.IsValid(req) {
			next.ServeHTTP(res, req)
		} else {
			http.Redirect(res, req, redirect, http.StatusFound)
		}
	})
}

// IsValid checks if the user request is authenticated
func (a *Auth) IsValid(req *http.Request) bool {
	_, err := token.GetToken(req)

	return err == nil
}

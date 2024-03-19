package auth

import (
	"github.com/nilslice/jwt"
	"net/http"
)

type Auth struct {
}

// Check is HTTP middleware to ensure the request has proper token credentials
func (a *Auth) Check(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		redir := req.URL.Scheme + req.URL.Host + "/admin/login"

		if a.IsValid(req) {
			next.ServeHTTP(res, req)
		} else {
			http.Redirect(res, req, redir, http.StatusFound)
		}
	})
}

// IsValid checks if the user request is authenticated
func (a *Auth) IsValid(req *http.Request) bool {
	// check if token exists in cookie
	cookie, err := req.Cookie("_token")
	if err != nil {
		return false
	}
	// validate it and allow or redirect request
	token := cookie.Value
	return jwt.Passes(token)
}

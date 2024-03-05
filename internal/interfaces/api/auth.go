package api

import (
	"github.com/nilslice/jwt"
	"net/http"
)

// Auth is HTTP middleware to ensure the request has proper token credentials
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		redir := req.URL.Scheme + req.URL.Host + "/admin/login"

		if IsValid(req) {
			next.ServeHTTP(res, req)
		} else {
			http.Redirect(res, req, redir, http.StatusFound)
		}
	})
}

// IsValid checks if the user request is authenticated
func IsValid(req *http.Request) bool {
	// check if token exists in cookie
	cookie, err := req.Cookie("_token")
	if err != nil {
		return false
	}
	// validate it and allow or redirect request
	token := cookie.Value
	return jwt.Passes(token)
}

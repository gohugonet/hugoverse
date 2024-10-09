package token

import (
	"errors"
	"github.com/nilslice/jwt"
	"net/http"
	"strings"
)

func GetEmail(req *http.Request) (string, error) {
	token, err := GetToken(req)
	if err != nil {
		return "", err
	}
	claims := parseToken(token)

	return claims[userKey].(string), nil
}

func GetToken(req *http.Request) (string, error) {
	// check if token exists in cookie
	cookie, err := req.Cookie("_token")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		return "", err
	}

	if err == nil {
		if jwt.Passes(cookie.Value) {
			return cookie.Value, nil
		}
	}

	authHeader := req.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if jwt.Passes(authHeader) {
		return authHeader, nil
	}

	return "", errors.New("token not found in cookie or header")
}

func parseToken(token string) map[string]interface{} {
	return jwt.GetClaims(token)
}

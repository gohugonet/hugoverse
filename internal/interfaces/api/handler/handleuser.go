package handler

import (
	"encoding/json"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/token"
	"net/http"
	"strings"
)

func (s *Handler) UserRegisterHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			s.log.Errorf("Error parsing login form: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// check email & password
		email := strings.ToLower(req.FormValue("email"))
		pwd := req.FormValue("password")

		found := s.adminApp.IsUserExists(email)
		s.log.Println("[UserRegisterHandler]: ", email, found)

		if found {
			s.log.Errorf("User already exists: %v", email)
			res.WriteHeader(http.StatusConflict)
			return
		}

		_, err = s.adminApp.NewUser(email, pwd)
		if err != nil {
			s.log.Errorf("Error creating new user: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		nt, _, err := token.New(email)
		if err != nil {
			s.log.Errorf("Error creating new token: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tokenJSON, err := json.Marshal(nt)
		if err != nil {
			s.log.Errorf("Error marshalling token: %v", err)
			return
		}

		j, err := s.res.FmtJSON(tokenJSON)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusCreated)
		s.res.Json(res, j)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Handler) UserLoginHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			s.log.Errorf("Error parsing login form: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// check email & password
		email := strings.ToLower(req.FormValue("email"))
		pwd := req.FormValue("password")

		err = s.adminApp.ValidateUser(email, pwd)
		if err != nil {
			s.log.Errorf("Error validating user: %v", err)
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		nt, _, err := token.New(email)
		if err != nil {
			s.log.Errorf("Error creating new token: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		tokenJSON, err := json.Marshal(nt)
		if err != nil {
			s.log.Errorf("Error marshalling token: %v", err)
			return
		}

		j, err := s.res.FmtJSON(tokenJSON)
		if err != nil {
			s.log.Errorf("Error formatting JSON: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusCreated)
		s.res.Json(res, j)

	default:
		s.log.Errorf("Method not allowed: %s", req.Method)
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

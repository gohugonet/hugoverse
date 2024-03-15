package api

import (
	"encoding/base64"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/admin"
	"github.com/gohugonet/hugoverse/pkg/db"
	"github.com/nilslice/jwt"
	"log"
	"net/http"
	"strings"
	"time"
)

func (s *Server) registerAdminHandler() {
	s.mux.HandleFunc("/admin", Auth(s.adminHandler))

	s.mux.HandleFunc("/admin/login", s.loginHandler)
	s.mux.HandleFunc("/admin/logout", s.logoutHandler)

	s.mux.HandleFunc("/admin/configure", Auth(s.configHandler))

	s.mux.HandleFunc("/admin/contents", Auth(s.contentsHandler))
	s.mux.HandleFunc("/admin/contents/search", Auth(s.searchHandler))

	s.mux.HandleFunc("/admin/edit", Auth(s.editHandler))
	s.mux.HandleFunc("/admin/edit/delete", Auth(s.deleteHandler))

	s.mux.HandleFunc("/admin/uploads", Auth(s.uploadContentsHandler))
	s.mux.HandleFunc("/admin/uploads/search", Auth(s.uploadSearchHandler))
	s.mux.HandleFunc("/admin/edit/upload", Auth(s.editUploadHandler))
	s.mux.HandleFunc("/admin/edit/upload/delete", Auth(s.deleteUploadHandler))

	s.mux.HandleFunc("/admin/init", s.initHandler)

	staticDir := adminStaticDir()
	s.mux.Handle("/admin/static/", s.CacheControl(
		http.StripPrefix("/admin/static",
			http.FileServer(restrict(http.Dir(staticDir))))))

	uploadsDir := uploadDir()
	s.mux.Handle("/api/uploads/", Record(s.CORS(s.CacheControl(
		http.StripPrefix("/api/uploads/",
			http.FileServer(restrict(http.Dir(uploadsDir))))))))

}

func (s *Server) adminHandler(res http.ResponseWriter, req *http.Request) {
	view, err := s.adminView.Dashboard()
	if err != nil {
		s.Log.Errorf("Error rendering admin view: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "text/html")
	res.Write(view)
}

func (s *Server) initHandler(res http.ResponseWriter, req *http.Request) {
	if db.SystemInitComplete() {
		http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
		return
	}

	switch req.Method {
	case http.MethodGet:
		view, err := admin.SetupView(s.adminApp.Name())
		if err != nil {
			s.Log.Errorf("Error rendering admin view: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "text/html")
		res.Write(view)
	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			s.Log.Errorf("Error parsing form: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// get the site name from post to encode and use as secret
		name := []byte(req.FormValue("name") + s.adminApp.NewETage())
		secret := base64.StdEncoding.EncodeToString(name)
		req.Form.Set("client_secret", secret)

		// generate an Etag to use for response caching
		etag := s.adminApp.NewETage()
		req.Form.Set("etag", etag)

		// create and save admin user
		email := strings.ToLower(req.FormValue("email"))
		password := req.FormValue("password")

		_, err = s.adminApp.NewUser(email, password)
		if err != nil {
			s.Log.Errorf("Error creating new user: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// set HTTP port which should be previously added to config cache
		req.Form.Set("http_port", s.adminApp.HttpPort())

		// set initial user email as admin_email and make config
		req.Form.Set("admin_email", email)

		err = s.adminApp.SetConfig(req.Form)
		if err != nil {
			s.Log.Errorf("Error setting config: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// add _token cookie for login persistence
		week := time.Now().Add(time.Hour * 24 * 7)
		claims := map[string]interface{}{
			"exp":  week.Unix(),
			"user": email,
		}

		jwt.Secret([]byte(secret))
		token, err := jwt.New(claims)
		if err != nil {
			s.Log.Errorf("Error creating JWT: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.SetCookie(res, &http.Cookie{
			Name:    "_token",
			Value:   token,
			Expires: week,
			Path:    "/",
		})

		redir := strings.TrimSuffix(req.URL.String(), "/init")
		http.Redirect(res, req, redir, http.StatusFound)
	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) loginHandler(res http.ResponseWriter, req *http.Request) {
	if !db.SystemInitComplete() {
		redir := req.URL.Scheme + req.URL.Host + "/admin/init"
		http.Redirect(res, req, redir, http.StatusFound)
		return
	}

	switch req.Method {
	case http.MethodGet:
		if IsValid(req) {
			http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			return
		}

		view, err := admin.Login(s.adminApp.Name())
		if err != nil {
			s.Log.Errorf("Error rendering login view: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(view)

	case http.MethodPost:
		if IsValid(req) {
			http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin", http.StatusFound)
			return
		}

		err := req.ParseForm()
		if err != nil {
			s.Log.Errorf("Error parsing login form: %v", err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		// check email & password
		email := strings.ToLower(req.FormValue("email"))
		pwd := req.FormValue("password")

		err = s.adminApp.ValidateUser(email, pwd)
		if err != nil {
			s.Log.Errorf("Error validating user: %v", err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		// create new token
		week := time.Now().Add(time.Hour * 24 * 7)
		claims := map[string]interface{}{
			"exp":  week,
			"user": email,
		}
		token, err := jwt.New(claims)
		if err != nil {
			s.Log.Errorf("Error creating JWT: %v", err)
			http.Redirect(res, req, req.URL.String(), http.StatusFound)
			return
		}

		// add it to cookie +1 week expiration
		http.SetCookie(res, &http.Cookie{
			Name:    "_token",
			Value:   token,
			Expires: week,
			Path:    "/",
		})

		http.Redirect(res, req, strings.TrimSuffix(req.URL.String(), "/login"), http.StatusFound)
	}
}

func (s *Server) logoutHandler(res http.ResponseWriter, req *http.Request) {
	http.SetCookie(res, &http.Cookie{
		Name:    "_token",
		Expires: time.Unix(0, 0),
		Value:   "",
		Path:    "/",
	})

	http.Redirect(res, req, req.URL.Scheme+req.URL.Host+"/admin/login", http.StatusFound)
}

func (s *Server) configHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		cfg, err := s.adminApp.ConfigEditor()
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		adminView, err := s.adminView.SubView(cfg)
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/html")
		res.Write(adminView)

	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			log.Println(err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = s.adminApp.SetConfig(req.Form)
		if err != nil {
			s.Log.Errorf("Error setting config: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(res, req, req.URL.String(), http.StatusFound)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}

}

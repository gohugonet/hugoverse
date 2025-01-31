package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/application"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/form"
	"github.com/gohugonet/hugoverse/pkg/fs/static"
	"github.com/gohugonet/hugoverse/pkg/rand"
	"github.com/spf13/afero"
	"log"
	"net/http"
	"path"
)

func (s *Handler) PreviewContentHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	id := q.Get("id")
	t := q.Get("type")
	status := q.Get("status")

	if t == "" || id == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err := req.ParseMultipartForm(form.MaxMemory)
	if err != nil {
		s.log.Errorf("Error parsing deploy form: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	pt, ok := s.contentApp.GetContentCreator(t)
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	p := pt()
	_, ok = p.(content.Deployable)
	if !ok {
		log.Println("[Response] error: Type", t, "does not implement item.Deployable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	t, err = s.contentApp.BuildTarget(t, id, status)
	if err != nil {
		s.log.Errorf("Error building: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = application.GenerateStaticSiteWithTarget(t)
	if err != nil {
		s.log.Errorf("Error building site %s for deployment with error : %v", id, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	d := &valueobject.Domain{
		Root:  "app.mdfriday.com",
		Sub:   fmt.Sprintf("%s-%s", "mdf", rand.ShortString(6)),
		Owner: "MDFriday",
	}

	sd := &valueobject.Deployment{
		SiteName: d.Sub,
		HostName: "Netlify",
		Status:   "pending",
	}

	err = application.DeployToNetlify(t, sd, d, s.adminApp.Netlify.Token())
	if err != nil {
		s.log.Errorf("Error building: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.contentApp.UpdateContentObject(
		&valueobject.Preview{
			Domain:   sd.Domain,
			SiteID:   sd.SiteID,
			SiteName: sd.SiteName,
			SitePath: sd.SitePath,
			HostName: sd.HostName,
			Status:   sd.Status,
		}); err != nil {
		s.log.Errorf("Error updating deployment: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal("https://" + d.FullDomain())
	if err != nil {
		s.log.Errorf("Error marshalling token: %v", err)
		return
	}

	j, err := s.res.FmtJSON(jsonBytes)
	if err != nil {
		s.log.Errorf("Error formatting JSON: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	s.res.Json(res, j)
}

func (s *Handler) PreviewContentHandlerLocal(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	id := q.Get("id")
	t := q.Get("type")
	status := q.Get("status")

	if t == "" || id == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	pt, ok := s.contentApp.GetContentCreator(t)
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	p := pt()
	_, ok = p.(content.Buildable)
	if !ok {
		s.log.Println("[Response] error: Type", t, "does not implement item.Buildable or embed item.Item.")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	t, pf, err := s.contentApp.PreviewTarget(t, id, status)
	if err != nil {
		s.log.Errorf("Error preview for site %s with error: %v", id, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = application.GenerateStaticSiteWithTarget(t)
	if err != nil {
		s.log.Errorf("Error building site %s for preview with error : %v", id, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	osFs := afero.NewOsFs()
	prefs := afero.NewBasePathFs(osFs, path.Join(application.PreviewDir(), pf))
	pubfs := afero.NewBasePathFs(osFs, path.Join(t, "public"))

	if err := static.Copy([]afero.Fs{pubfs}, prefs); err != nil {
		s.log.Errorf("Error copying site %s for preview with error : %v", id, err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenJSON, err := json.Marshal(path.Join(getRootURL(req), application.PreviewFolder(), pf))
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

	res.WriteHeader(http.StatusOK)
	s.res.Json(res, j)
}

func getRootURL(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, req.Host)
}

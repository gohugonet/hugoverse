package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/application"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"log"
	"net/http"
)

func (s *Handler) DeployContentHandler(res http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	id := q.Get("id")
	t := q.Get("type")
	status := q.Get("status")

	if t == "" || id == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err := req.ParseForm()
	if err != nil {
		s.log.Errorf("Error parsing deploy form: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	netlify := req.PostForm.Get("netlify")
	root := req.PostForm.Get("domain")
	if netlify == "" || root == "" {
		netlify = s.adminApp.Netlify.Token()
		root = "app.mdfriday.com"
	}

	d, err := s.contentApp.ApplyDomain(id, root)
	if err != nil {
		s.log.Errorf("Error applying domain: %v", err)
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

	sd, err := s.contentApp.GetDeployment(id, d)
	if err != nil {
		s.log.Errorf("Error getting deployment: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = application.DeployToNetlify(t, sd, netlify)
	if err != nil {
		s.log.Errorf("Error building: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := s.contentApp.UpdateContentObject(sd); err != nil {
		s.log.Errorf("Error updating deployment: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(fmt.Sprintf("https://%s.app.mdfriday.com", sd.Domain))
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

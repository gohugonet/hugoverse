package handler

import (
	"encoding/json"
	"github.com/gohugonet/hugoverse/internal/application"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/form"
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

	err := req.ParseMultipartForm(form.MaxMemory)
	if err != nil {
		s.log.Errorf("Error parsing deploy form: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	hostName := req.FormValue("host_name")
	hostToken := req.FormValue("host_token")
	root := req.FormValue("domain")

	if hostToken == "" || root == "" {
		s.log.Errorf("Both host_token and domain must be set")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	d, isTaken, err := s.contentApp.ApplyDomain(id, root)
	if !isTaken && err != nil {
		s.log.Errorf("Error applying domain: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	} else if isTaken {
		s.log.Errorf("Domain already taken: %s", err.Error())
		res.WriteHeader(http.StatusConflict)
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

	var target string

	sc, err := s.contentApp.GetContentObject(t, id)
	if err != nil {
		s.log.Errorf("Error getting deploy content: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if site, ok := sc.(*valueobject.Site); ok {
		target = site.WorkingDir
	}

	if target == "" {
		target, err = s.contentApp.BuildTarget(t, id, status)
		if err != nil {
			s.log.Errorf("Error building: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = application.GenerateStaticSiteWithTarget(target)
		if err != nil {
			s.log.Errorf("Error building site %s for deployment with error : %v", id, err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	sd, err := s.contentApp.GetDeployment(d, hostName)
	if err != nil {
		s.log.Errorf("Error getting deployment: %v", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if hostName != "Netlify" {
		s.log.Errorf("Error: Netlify only supported for now")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = application.DeployToNetlify(target, sd, d, hostToken)
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

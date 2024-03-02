package api

import "net/http"

type demo struct {
	Name string
}

func (s *Server) handleDemo(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	s.writeJSONResponse(w, &demo{Name: "demo"}, http.StatusOK)
}

package query

import (
	"net/http"
	"strings"
)

func Order(req *http.Request) string {
	q := req.URL.Query()

	order := strings.ToLower(q.Get("order"))
	if order != "asc" {
		order = "desc"
	}

	return order
}

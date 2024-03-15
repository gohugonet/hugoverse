package query

import (
	"net/http"
	"strconv"
)

func Count(req *http.Request) (int, error) {
	q := req.URL.Query()

	count, err := strconv.Atoi(q.Get("count")) // int: determines number of posts to return (10 default, -1 is all)
	if err != nil {
		if q.Get("count") == "" {
			count = 10
			return count, nil
		}
	}

	return count, err
}

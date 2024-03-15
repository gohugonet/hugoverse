package query

import (
	"net/http"
	"strconv"
)

func Offset(req *http.Request) (int, error) {
	q := req.URL.Query()

	offset, err := strconv.Atoi(q.Get("offset")) // int: multiplier of count for pagination (0 default)
	if err != nil {
		if q.Get("offset") == "" {
			offset = 0
			return offset, nil
		}
	}
	return offset, err
}

package handler

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/content"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"log"
	"net/http"
)

func omit(res http.ResponseWriter, req *http.Request, it interface{}, data []byte) ([]byte, error) {
	// is it Omittable
	om, ok := it.(content.Omittable)
	if !ok {
		return data, nil
	}

	return omitFields(res, req, om, data, "data")
}

func omitFields(res http.ResponseWriter, req *http.Request, om content.Omittable, data []byte, pathPrefix string) ([]byte, error) {
	// get fields to omit from json data
	fields, err := om.Omit(res, req)
	if err != nil {
		return nil, err
	}

	// remove each field from json, all responses contain json object(s) in top-level "data" array
	n := int(gjson.GetBytes(data, pathPrefix+".#").Int())
	for i := 0; i < n; i++ {
		for k := range fields {
			var err error
			data, err = sjson.DeleteBytes(data, fmt.Sprintf("%s.%d.%s", pathPrefix, i, fields[k]))
			if err != nil {
				log.Println("Erorr omitting field:", fields[k], "from item.Omittable:", om)
				return nil, err
			}
		}
	}

	return data, nil
}

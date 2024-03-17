package editor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/db"
	"html"
	"html/template"
	"log"
)

// RefSelect returns the []byte of a <select> HTML element plus internal <options> with a label.
// IMPORTANT:
// The `fieldName` argument will cause a panic if it is not exactly the string
// form of the struct field that this editor input is representing
func RefSelect(fieldName string, p interface{}, attrs map[string]string, contentType, tmplString string) []byte {
	options, err := encodeDataToOptions(contentType, tmplString)
	if err != nil {
		log.Println("Error encoding data to options for", contentType, err)
		return nil
	}

	return Select(fieldName, p, attrs, options)
}

func encodeDataToOptions(contentType, tmplString string) (map[string]string, error) {
	// encode all content type from db into options map
	// options in form of map["/api/content?type=<contentType>&id=<id>"]t.String()
	options := make(map[string]string)

	data := db.ContentAll(contentType)

	// make template for option html display
	tmpl := template.Must(template.New(contentType).Parse(tmplString))

	for _, jsonData := range data {
		var item map[string]interface{}

		err := json.Unmarshal(jsonData, &item)
		if err != nil {
			return nil, err
		}

		k := fmt.Sprintf("/api/content?type=%s&id=%.0f", contentType, item["id"].(float64))
		v := &bytes.Buffer{}
		err = tmpl.Execute(v, item)
		if err != nil {
			return nil, fmt.Errorf(
				"error executing template for reference of %s: %s",
				contentType, err.Error())
		}

		options[k] = html.UnescapeString(v.String())
	}

	return options, nil
}

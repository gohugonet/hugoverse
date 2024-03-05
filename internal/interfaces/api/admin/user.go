package admin

import (
	"bytes"
	"html/template"
)

// Login ...
func Login(name string) ([]byte, error) {
	html := startAdminHTML + loginAdminHTML + endAdminHTML

	a := admin{
		Logo: name,
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("login").Parse(html))
	err := tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

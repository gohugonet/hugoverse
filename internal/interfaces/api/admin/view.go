package admin

import (
	"bytes"
	"html/template"
)

// SetupView ...
func SetupView(name string) ([]byte, error) {
	html := startAdminHTML + initAdminHTML + endAdminHTML

	a := admin{
		Logo: name,
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("init").Parse(html))
	err := tmpl.Execute(buf, a)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type admin struct {
	Logo    string
	Types   map[string]func() interface{}
	Subview template.HTML
}

// Admin ...
func Admin(view []byte, name string, ts map[string]func() interface{}) (_ []byte, err error) {
	a := admin{
		Logo:    name,
		Types:   ts,
		Subview: template.HTML(view),
	}

	buf := &bytes.Buffer{}
	html := startAdminHTML + mainAdminHTML + endAdminHTML
	tmpl := template.Must(template.New("admin").Parse(html))
	err = tmpl.Execute(buf, a)
	if err != nil {
		return
	}

	return buf.Bytes(), nil
}

package admin

import (
	"bytes"
	"github.com/gohugonet/hugoverse/internal/interfaces/api/analytics"
	"html/template"
)

// Dashboard returns the admin view with analytics dashboard
func Dashboard(name string, ts map[string]func() interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	data, err := analytics.ChartData()
	if err != nil {
		return nil, err
	}

	tmpl := template.Must(template.New("analytics").Parse(analyticsHTML))
	err = tmpl.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return Admin(buf.Bytes(), name, ts)
}

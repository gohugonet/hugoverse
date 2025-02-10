package admin

import (
	"bytes"
	"github.com/mdfriday/hugoverse/internal/interfaces/api/record/analytics"
	"html/template"
)

// Dashboard returns the AdminView view with analytics dashboard
func (v *View) Dashboard() ([]byte, error) {
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
	return v.SubView(buf.Bytes())
}

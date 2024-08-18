package testkit

const indexTemplateContent = `
<head>
  {{ partial "doc/head.html" . }}
</head>
<body>{{ .Content }}</body>

`

const headTemplateContent = `
<div class="flex align-center justify-between">
  <strong>{{ .Title }}</strong>
</div>

`

type TemplateIndex struct {
	Title   string
	Content string
}

func NewTemplateIndex(title string, content string) *TemplateIndex {
	return &TemplateIndex{
		Title:   title,
		Content: content,
	}
}

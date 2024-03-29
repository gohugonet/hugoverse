package valueobject

type TemplateDescriptor struct {
	name      string
	extension string
}

func NewTemplateDescriptor(name, ext string) *TemplateDescriptor {
	return &TemplateDescriptor{
		name:      name,
		extension: ext,
	}
}

func (td *TemplateDescriptor) Name() string {
	return td.name
}

func (td *TemplateDescriptor) Extension() string {
	return td.extension
}

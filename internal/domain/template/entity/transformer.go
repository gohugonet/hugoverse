package entity

import "github.com/gohugonet/hugoverse/internal/domain/template/valueobject"

type AstTransformer struct {
	// Holds name and source of template definitions not found during the first
	// AST transformation pass.
	TransformNotFound map[string]*valueobject.State
}

func (t *AstTransformer) applyTemplateTransformers(ns *Namespace, ts *valueobject.State) (*valueobject.Context, error) {
	c, err := valueobject.ApplyTemplateTransformers(ts, ns.newTemplateLookup(ts))
	if err != nil {
		return nil, err
	}

	for k := range c.TemplateNotFound {
		t.TransformNotFound[k] = ts
	}

	return c, err
}

func (t *AstTransformer) post(ns *Namespace) error {
	for name, source := range t.TransformNotFound {
		lookup := ns.newTemplateLookup(source)
		templ := lookup(name)
		if templ != nil {
			_, err := valueobject.ApplyTemplateTransformers(templ, lookup)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

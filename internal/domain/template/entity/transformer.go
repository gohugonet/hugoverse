package entity

import "github.com/gohugonet/hugoverse/internal/domain/template/valueobject"

type templateLookupFunc func(in *valueobject.State) func(name string) *valueobject.State

type AstTransformer struct {
	// Holds name and source of template definitions not found during the first
	// AST transformation pass.
	TransformNotFound map[string]*valueobject.State
}

func (t *AstTransformer) applyTemplateTransformers(lookup templateLookupFunc, ts *valueobject.State) (*Context, error) {
	c, err := ApplyTemplateTransformers(ts, lookup(ts))
	if err != nil {
		return nil, err
	}

	for k := range c.TemplateNotFound {
		t.TransformNotFound[k] = ts
	}

	return c, err
}

func (t *AstTransformer) post(lu templateLookupFunc) error {
	for name, source := range t.TransformNotFound {
		lookup := lu(source)
		templ := lookup(name)
		if templ != nil {
			_, err := ApplyTemplateTransformers(templ, lookup)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

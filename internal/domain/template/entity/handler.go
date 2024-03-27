package entity

import (
	"fmt"
	fsFactory "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/spf13/afero"
	"path/filepath"
	"strings"
)

type TemplateHandler struct {
	Main     *TemplateNamespace
	LayoutFs afero.Fs
}

func (t *TemplateHandler) LoadTemplates() error {
	walker := func(path string, fi fsVO.FileMetaInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}

		name := strings.TrimPrefix(filepath.ToSlash(path), "/")
		fmt.Println(">>> add template: ", path, name)

		if err := t.addTemplateFile(name, path); err != nil {
			return err
		}

		return nil
	}

	return fsFactory.NewWalkway(t.LayoutFs, "", walker).Walk()
}

func (t *TemplateHandler) addTemplateFile(name, path string) error {
	getTemplate := func(filename string) (templateInfo, error) {
		afs := t.LayoutFs
		b, err := afero.ReadFile(afs, filename)
		if err != nil {
			return templateInfo{filename: filename, fs: afs}, err
		}

		return templateInfo{
			name:     name,
			template: string(b),
			filename: filename,
			fs:       afs,
		}, nil
	}

	tinfo, err := getTemplate(path)
	if err != nil {
		return err
	}

	_, err = t.addTemplateTo(tinfo, t.Main)
	if err != nil {
		return tinfo.errWithFileContext("parse failed", err)
	}

	return nil
}

func (t *TemplateHandler) addTemplateTo(info templateInfo, to *TemplateNamespace) (*TemplateState, error) {
	return to.parse(info)
}

func (t *TemplateHandler) Lookup(name string) (template.Template, bool) {
	tmpl, found := t.Main.Lookup(name)
	if found {
		return tmpl, true
	}

	return nil, false
}

func (t *TemplateHandler) LookupLayout(layouts []string) (template.Template, bool, error) {
	return t.findLayout(layouts)
}

func (t *TemplateHandler) findLayout(layouts []string) (template.Template, bool, error) {
	for _, name := range layouts {
		templ, found := t.Main.Lookup(name)
		if found {
			return templ, true, nil
		}
	}

	return nil, false, nil
}

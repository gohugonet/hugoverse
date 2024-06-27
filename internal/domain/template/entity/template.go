package entity

import (
	"fmt"
	fsFactory "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"path/filepath"
	"strings"
	"sync"
)

type Template struct {
	*Executor
	*Lookup

	Parser *Parser

	Main *Namespace
	Fs   template.Fs

	shortcodeOnce sync.Once
	*Shortcode
}

func (t *Template) LookupLayout(d template.LayoutDescriptor) (template.Preparer, bool, error) {
	p, found, err := t.Lookup.lookupLayout(d, t.Main)
	if err != nil {
		return nil, false, err
	}

	if found {
		return p, true, nil
	}

	overlay, base, found := t.Lookup.findLayoutInfo(d)
	if found {
		ts, found, err := t.Parser.ParseOverlap(overlay, base)
		if found {
			if err = t.Parser.Transform(t.Main, ts); err != nil {
				fmt.Printf("LookupLayout 3 %+v, %v, %v--- ", ts, err, found)
			}

			if err := t.extractPartials(ts.Preparer); err != nil {
				return nil, false, err
			}
			return ts.Preparer, found, nil
		}
	}

	return nil, false, nil
}

func (t *Template) LoadTemplates() error {
	walker := func(path string, fi fsVO.FileMetaInfo, err error) error {
		if err != nil {
			fmt.Println("LoadTemplates --- ", path, err)
		}

		if fi.IsDir() {
			return nil
		}

		if isDotFile(path) || isBackupFile(path) {
			return nil
		}

		name := strings.TrimPrefix(filepath.ToSlash(path), "/")

		if err := t.addTemplateFile(name, fi); err != nil {
			return err
		}

		return nil
	}

	if err := fsFactory.NewWalkway(t.Fs.LayoutFs(), "", walker).Walk(); err != nil {
		if !herrors.IsNotExist(err) {
			return err
		}
		return nil
	}

	return nil
}

func (t *Template) addTemplateFile(name string, fim fsVO.FileMetaInfo) error {
	tinfo, err := valueobject.LoadTemplate(name, fim)
	if err != nil {
		return err
	}

	if t.Lookup.BaseOf.IsBaseTemplatePath(name) {
		t.Lookup.BaseOf.AddBaseOf(name, tinfo)
		return nil
	}

	if t.Lookup.BaseOf.NeedsBaseOf(name, tinfo.Template) {
		t.Lookup.BaseOf.AddNeedsBaseOf(name, tinfo)
		return nil
	}

	state, err := t.Parser.Parse(tinfo)
	if err != nil {
		return tinfo.ErrWithFileContext("parse failed", err)
	}

	t.Main.addTemplate(tinfo.Name, state)

	if err := t.Parser.Transform(t.Main, state); err != nil {
		fmt.Println(tinfo.ErrWithFileContext("ast transform parse failed", err))
	}

	return nil
}

func isDotFile(path string) bool {
	return filepath.Base(path)[0] == '.'
}

func isBackupFile(path string) bool {
	return path[len(path)-1] == '~'
}

func (t *Template) PostTransform() error {
	defineCheckedHTML := false
	defineCheckedText := false

	for _, v := range t.Main.Templates {
		if v.Typ == template.TypeShortcode {
			t.getShortcode().addShortcodeVariant(v)
		}

		if defineCheckedHTML && defineCheckedText {
			continue
		}

		isText := isText(v.Preparer)
		if isText {
			if defineCheckedText {
				continue
			}
			defineCheckedText = true
		} else {
			if defineCheckedHTML {
				continue
			}
			defineCheckedHTML = true
		}

		if err := t.extractPartials(v.Preparer); err != nil {
			return err
		}
	}

	if err := t.Parser.Ast.post(t.Main.newTemplateLookup); err != nil {
		return err
	}

	return nil
}

func (t *Template) getShortcode() *Shortcode {
	t.shortcodeOnce.Do(func() {
		t.Shortcode = &Shortcode{
			shortcodes: map[string]*shortcodeTemplates{},
		}
	})
	return t.Shortcode
}

func isText(templ template.Preparer) bool {
	_, isText := templ.(*texttemplate.Template)
	return isText
}

func (t *Template) extractPartials(templ template.Preparer) error {
	partials := t.Main.getUnregisteredPartials(templ)
	for _, p := range partials {
		if err := t.Parser.Transform(t.Main, p); err != nil {
			return err
		}
		t.Main.addTemplate(p.Name(), p)
	}
	return nil
}

func (t *Template) Parse(name, tpl string) (template.Preparer, error) {
	return t.Parser.ParseWithLock(name, tpl)
}

package entity

import (
	"fmt"
	fsFactory "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	htmltemplate "github.com/gohugonet/hugoverse/pkg/template/htmltemplate"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	baseFileBase = "baseof"
)

type Template struct {
	*Executor
	*Lookup

	Ast *AstTransformer

	Main *Namespace
	Fs   template.Fs
}

func (t *Template) LookupLayout(d template.LayoutDescriptor) (template.Preparer, bool, error) {
	p, found, err := t.Lookup.lookupLayout(d)
	if err != nil {
		return nil, false, err
	}

	if found {
		return p, true, nil
	}

	ts, found, err := t.Lookup.findLayout(d)
	if found {
		_, err = t.Ast.applyTemplateTransformers(t.Main, ts)
		if err != nil {
			fmt.Printf("LookupLayout 3 %+v, %v, %v--- ", ts, err, found)
		}

		if err := t.extractPartials(ts.Preparer); err != nil {
			return nil, false, err
		}
		return ts.Preparer, found, nil
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
	getTemplate := func(fim fsVO.FileMetaInfo) (valueobject.Info, error) {
		meta := fim.Meta()
		f, err := meta.Open()
		if err != nil {
			return valueobject.Info{Meta: meta}, err
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return valueobject.Info{Meta: meta}, err
		}

		s := removeLeadingBOM(string(b))

		return valueobject.Info{
			Name:     name,
			IsText:   false,
			Template: s,
			Meta:     meta,
		}, nil
	}

	tinfo, err := getTemplate(fim)
	if err != nil {
		return err
	}

	if isBaseTemplatePath(name) {
		// Store it for later.
		t.Lookup.Baseof[name] = tinfo
		return nil
	}

	needsBaseof := !noBaseNeeded(name) && needsBaseTemplate(tinfo.Template)
	if needsBaseof {
		t.Lookup.NeedsBaseof[name] = tinfo
		return nil
	}

	state, err := t.addTemplateTo(tinfo, t.Main)
	if err != nil {
		return tinfo.ErrWithFileContext("parse failed", err)
	}
	_, err = t.Ast.applyTemplateTransformers(t.Main, state)
	if err != nil {
		fmt.Println(tinfo.ErrWithFileContext("ast transform parse failed", err))
	}

	return nil
}

func (t *Template) addTemplateTo(info valueobject.Info, to *Namespace) (*valueobject.State, error) {
	return to.parse(info)
}

func isDotFile(path string) bool {
	return filepath.Base(path)[0] == '.'
}

func isBackupFile(path string) bool {
	return path[len(path)-1] == '~'
}

func removeLeadingBOM(s string) string {
	const bom = '\ufeff'

	for i, r := range s {
		if i == 0 && r != bom {
			return s
		}
		if i > 0 {
			return s[i:]
		}
	}

	return s
}

func isBaseTemplatePath(path string) bool {
	return strings.Contains(filepath.Base(path), baseFileBase)
}

func noBaseNeeded(name string) bool {
	if strings.HasPrefix(name, "shortcodes/") || strings.HasPrefix(name, "partials/") {
		return true
	}
	return strings.Contains(name, "_markup/")
}

var baseTemplateDefineRe = regexp.MustCompile(`^{{-?\s*define`)

// needsBaseTemplate returns true if the first non-comment template block is a
// define block.
// If a base template does not exist, we will handle that when it's used.
func needsBaseTemplate(templ string) bool {
	idx := -1
	inComment := false
	for i := 0; i < len(templ); {
		if !inComment && strings.HasPrefix(templ[i:], "{{/*") {
			inComment = true
			i += 4
		} else if !inComment && strings.HasPrefix(templ[i:], "{{- /*") {
			inComment = true
			i += 6
		} else if inComment && strings.HasPrefix(templ[i:], "*/}}") {
			inComment = false
			i += 4
		} else if inComment && strings.HasPrefix(templ[i:], "*/ -}}") {
			inComment = false
			i += 6
		} else {
			r, size := utf8.DecodeRuneInString(templ[i:])
			if !inComment {
				if strings.HasPrefix(templ[i:], "{{") {
					idx = i
					break
				} else if !unicode.IsSpace(r) {
					break
				}
			}
			i += size
		}
	}

	if idx == -1 {
		return false
	}

	return baseTemplateDefineRe.MatchString(templ[idx:])
}

func (t *Template) PostTransform() error {
	defineCheckedHTML := false
	defineCheckedText := false

	for _, v := range t.Main.Templates {
		if v.Typ == template.TypeShortcode {
			panic("not implemented for shortcode in PostTransform")
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

	if err := t.Ast.post(t.Main); err != nil {
		return err
	}

	return nil
}

func isText(templ template.Preparer) bool {
	_, isText := templ.(*texttemplate.Template)
	return isText
}

func (t *Template) extractPartials(templ template.Preparer) error {
	templs := templates(templ)
	for _, tmpl := range templs {
		if tmpl.Name() == "" || !strings.HasPrefix(tmpl.Name(), "partials/") {
			continue
		}

		ts := newTemplateState(tmpl, valueobject.Info{Name: tmpl.Name()}, nil)
		ts.Typ = template.TypePartial

		if err := t.Main.add(tmpl, ts); err != nil {
			return err
		}
	}

	return nil
}

func templates(in template.Preparer) []template.Preparer {
	var templs []template.Preparer
	in = unwrap(in)
	if textt, ok := in.(*texttemplate.Template); ok {
		for _, t := range textt.Templates() {
			templs = append(templs, t)
		}
	}

	if htmlt, ok := in.(*htmltemplate.Template); ok {
		for _, t := range htmlt.Templates() {
			templs = append(templs, t)
		}
	}

	return templs
}

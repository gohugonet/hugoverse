package entity

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/internal/domain/template/valueobject"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	texttemplate "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	iofs "io/fs"
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

	Log loggers.Logger

	LayoutTemplateCache   map[string]valueobject.LayoutCacheEntry
	layoutTemplateCacheMu sync.RWMutex
}

func (t *Template) MarkReady() error {
	return t.Parser.MarkReady()
}

func (t *Template) LookupLayout(names []string) (template.Preparer, bool, error) {
	cacheKey := valueobject.LayoutCacheKey{Names: names}
	if cacheKey.IsEmpty() {
		t.Log.Warnf("LookupLayout called with empty names")

		return nil, false, nil
	}

	key := cacheKey.String()
	t.layoutTemplateCacheMu.RLock()
	if cacheVal, found := t.LayoutTemplateCache[key]; found {
		t.layoutTemplateCacheMu.RUnlock()
		return cacheVal.Templ, cacheVal.Found, cacheVal.Err
	}
	t.layoutTemplateCacheMu.RUnlock()

	t.layoutTemplateCacheMu.Lock()
	defer t.layoutTemplateCacheMu.Unlock()

	p, found, err := t.Lookup.findStandalone(names, t.Main)
	if err != nil {
		return nil, false, err
	}

	if found {
		cacheVal := valueobject.LayoutCacheEntry{Found: found, Templ: p, Err: nil}
		t.LayoutTemplateCache[key] = cacheVal

		return p, true, nil
	}

	overlay, base, found := t.Lookup.findDependentInfo(names)
	if found {
		ts, found, err := t.Parser.ParseOverlap(overlay, base)
		if found {
			if err = t.Parser.Transform(t.Main, ts); err != nil {
				t.Log.Printf("LookupLayout transform %+v, %v, %v--- ", ts, err, found)
			}

			if err := t.extractPartials(ts.Preparer); err != nil {
				return nil, false, err
			}

			cacheVal := valueobject.LayoutCacheEntry{Found: found, Templ: ts.Preparer, Err: nil}
			t.LayoutTemplateCache[key] = cacheVal

			return ts.Preparer, found, nil
		}
	}

	return nil, false, nil
}

//go:embed all:embedded/templates/*
//go:embed embedded/templates/_default/*
//go:embed embedded/templates/_server/*
var embeddedTemplatesFs embed.FS
var embeddedTemplatesAliases = map[string][]string{
	"shortcodes/twitter.html": {"shortcodes/tweet.html"},
}

func (t *Template) LoadEmbedded() error {
	return iofs.WalkDir(embeddedTemplatesFs, ".", func(path string, d iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil || d.IsDir() {
			return nil
		}

		templb, err := embeddedTemplatesFs.ReadFile(path)
		if err != nil {
			return err
		}

		// Get the newlines on Windows in line with how we had it back when we used Go Generate
		// to write the templates to Go files.
		templ := string(bytes.ReplaceAll(templb, []byte("\r\n"), []byte("\n")))
		name := strings.TrimPrefix(filepath.ToSlash(path), "embedded/templates/")
		templateName := name

		// For the render hooks and the server templates it does not make sense to preserve the
		// double _internal double book-keeping,
		// just add it if its now provided by the user.
		if !strings.Contains(path, "_default/_markup") && !strings.HasPrefix(name, "_server/") && !strings.HasPrefix(name, "partials/_funcs/") {
			templateName = valueobject.InternalPathPrefix + name
		}

		if _, found := t.Main.findTemplate(templateName); !found {
			if err := t.addTemplateContent(valueobject.EmbeddedPathPrefix+templateName, templ); err != nil {
				return err
			}
		}

		if aliases, found := embeddedTemplatesAliases[name]; found {
			// TODO(bep) avoid reparsing these aliases
			for _, alias := range aliases {
				alias = valueobject.InternalPathPrefix + alias
				if err := t.addTemplateContent(valueobject.EmbeddedPathPrefix+alias, templ); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (t *Template) LoadTemplates() error {
	walker := func(path string, fi fs.FileMetaInfo) error {

		if fi.IsDir() {
			return nil
		}

		if isDotFile(path) || isBackupFile(path) {
			return nil
		}

		name := strings.TrimPrefix(filepath.ToSlash(path), "/")

		if err := t.addTemplateFileInfo(name, fi); err != nil {
			return err
		}

		return nil
	}

	if err := t.Fs.WalkLayouts("", fs.WalkCallback{
		HookPre:  nil,
		WalkFn:   walker,
		HookPost: nil,
	}, fs.WalkwayConfig{}); err != nil {
		if !herrors.IsNotExist(err) {
			return err
		}
	}

	return nil
}

func (t *Template) addTemplateFileInfo(name string, fim fs.FileMetaInfo) error {
	tinfo, err := valueobject.LoadTemplate(name, fim)
	if err != nil {
		return err
	}

	return t.addTemplate(tinfo.Name, tinfo)
}

func (t *Template) addTemplateContent(name, tpl string) error {
	tinfo, err := valueobject.LoadTemplateContent(name, tpl)
	if err != nil {
		return err
	}

	return t.addTemplate(tinfo.Name, tinfo)
}

func (t *Template) addTemplate(name string, tinfo valueobject.TemplateInfo) error {
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

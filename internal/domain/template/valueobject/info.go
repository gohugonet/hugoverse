package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"io"
	"strings"
)

const (
	textTmplNamePrefix = "_text/"

	shortcodesPathPrefix = "shortcodes/"
	TemplateVersion      = 2
)

var DefaultParseConfig = ParseConfig{
	Version: TemplateVersion,
}

var DefaultParseInfo = ParseInfo{
	Config: DefaultParseConfig,
}

func LoadTemplate(name string, fim valueobject.FileMetaInfo) (TemplateInfo, error) {
	meta := fim.Meta()
	f, err := meta.Open()
	if err != nil {
		return TemplateInfo{Meta: meta}, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return TemplateInfo{Meta: meta}, err
	}

	s := removeLeadingBOM(string(b))

	var isText bool
	name, isText = nameIsText(name)

	return TemplateInfo{
		Name:     name,
		IsText:   isText,
		Template: s,
		Meta:     meta,
	}, nil
}

func nameIsText(name string) (string, bool) {
	isText := strings.HasPrefix(name, textTmplNamePrefix)
	if isText {
		name = strings.TrimPrefix(name, textTmplNamePrefix)
	}
	return name, isText
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

type TemplateInfo struct {
	Name       string
	Template   string
	IsText     bool // HTML or plain text template.
	IsEmbedded bool

	Meta *valueobject.FileMeta
}

func (info TemplateInfo) IdentifierBase() string {
	return info.Name
}

func (info TemplateInfo) ErrWithFileContext(what string, err error) error {
	err = fmt.Errorf(what+": %w", err)
	fe := herrors.NewFileErrorFromName(err, info.Meta.Filename)
	f, err := info.Meta.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	return fe.UpdateContent(f, nil)
}

func (info TemplateInfo) ResolveType() template.Type {
	return resolveTemplateType(info.Name)
}

func (info TemplateInfo) IsZero() bool {
	return info.Name == ""
}

func resolveTemplateType(name string) template.Type {
	if isShortcode(name) {
		return template.TypeShortcode
	}

	if strings.Contains(name, "partials/") {
		return template.TypePartial
	}

	return template.TypeUndefined
}

func isShortcode(name string) bool {
	return strings.Contains(name, shortcodesPathPrefix)
}

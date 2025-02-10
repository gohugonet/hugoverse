package valueobject

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/fs"
	"github.com/mdfriday/hugoverse/internal/domain/template"
	"github.com/mdfriday/hugoverse/pkg/herrors"
	pio "github.com/mdfriday/hugoverse/pkg/io"
	"io"
	"strings"
)

var DefaultParseConfig = ParseConfig{
	Version: TemplateVersion,
}

var DefaultParseInfo = ParseInfo{
	Config: DefaultParseConfig,
}

func LoadTemplateContent(name string, content string) (TemplateInfo, error) {
	s := removeLeadingBOM(content)

	var isEmbedded bool
	if strings.HasPrefix(name, EmbeddedPathPrefix) {
		isEmbedded = true
		name = strings.TrimPrefix(name, EmbeddedPathPrefix)
	}

	var isText bool
	name, isText = nameIsText(name)

	return TemplateInfo{
		Name:       name,
		IsText:     isText,
		IsEmbedded: isEmbedded,
		Template:   s,
		Fi:         nil,
	}, nil
}

func LoadTemplate(name string, fim fs.FileMetaInfo) (TemplateInfo, error) {
	f, err := fim.Open()
	if err != nil {
		return TemplateInfo{Fi: fim}, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return TemplateInfo{Fi: fim}, err
	}

	s := removeLeadingBOM(string(b))

	var isEmbedded bool
	if strings.HasPrefix(name, EmbeddedPathPrefix) {
		isEmbedded = true
		name = strings.TrimPrefix(name, EmbeddedPathPrefix)
	}

	var isText bool
	name, isText = nameIsText(name)

	return TemplateInfo{
		Name:       name,
		IsText:     isText,
		IsEmbedded: isEmbedded,
		Template:   s,
		Fi:         fim,
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

	Fi fs.FileMetaInfo
}

func (info TemplateInfo) IdentifierBase() string {
	return info.Name
}

func (info TemplateInfo) ErrWithFileContext(what string, err error) error {
	err = fmt.Errorf(what+": %w", err)
	fe := herrors.NewFileErrorFromName(err, info.Name)
	f := pio.NewReadSeekerNoOpCloserFromString(info.Template)
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
	return strings.Contains(name, ShortcodesPathPrefix)
}

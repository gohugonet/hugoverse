package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/internal/domain/template"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"strings"
)

const (
	shortcodesPathPrefix = "shortcodes/"
	TemplateVersion      = 2
)

var DefaultParseConfig = ParseConfig{
	Version: TemplateVersion,
}

var DefaultParseInfo = ParseInfo{
	Config: DefaultParseConfig,
}

type Info struct {
	Name       string
	Template   string
	IsText     bool // HTML or plain text template.
	IsEmbedded bool

	Meta *valueobject.FileMeta
}

func (info Info) IdentifierBase() string {
	return info.Name
}

func (info Info) ErrWithFileContext(what string, err error) error {
	err = fmt.Errorf(what+": %w", err)
	fe := herrors.NewFileErrorFromName(err, info.Meta.Filename)
	f, err := info.Meta.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	return fe.UpdateContent(f, nil)
}

func (info Info) ResolveType() template.Type {
	return resolveTemplateType(info.Name)
}

func (info Info) IsZero() bool {
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

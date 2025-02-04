package template

import (
	"context"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/collections"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/compare"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/hugo"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/image"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/lang"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/os"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/resource"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/site"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/strings"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/transform"
	"github.com/gohugonet/hugoverse/pkg/template/funcs/urls"
	template "github.com/gohugonet/hugoverse/pkg/template/texttemplate"
	"io"
	"reflect"
)

type Type int

const (
	TypeUndefined Type = iota
	TypeShortcode
	TypePartial
)

type Fs interface {
	WalkLayouts(start string, cb fs.WalkCallback, conf fs.WalkwayConfig) error
}

type Service interface {
	Execute(ctx context.Context, name string, data any) (tmpl string, res string, err error)
}

type Template interface {
	Executor
	Lookup
}

type Executor interface {
	ExecuteWithContext(ctx context.Context, t Preparer, wr io.Writer, data any) error
}

type Lookup interface {
	LookupLayout(names []string) (Preparer, bool, error)
	GetFunc(name string) (reflect.Value, bool)
}

type Preparer interface {
	Name() string
	template.Preparer
}

type Identity interface {
	IdentifierBase() string
}

type Info interface {
	ParseInfo() ParseInfo
}

type ParseInfo interface {
	Return() bool
	Inner() bool
}

type CustomizedFunctions interface {
	transform.Markdown
	urls.URL
	urls.RefSource
	compare.TimeZone
	collections.Language
	strings.Title
	resource.Resource
	image.Image
	os.Os
	site.Service
	hugo.Info
	lang.Translator
}

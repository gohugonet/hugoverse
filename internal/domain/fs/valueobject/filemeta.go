package valueobject

import (
	"errors"
	"github.com/gohugonet/hugoverse/pkg/hreflect"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/spf13/afero"
	"reflect"
)

type MetaProvider interface {
	Meta() *FileMeta
}

type FileOpener func() (afero.File, error)

type FileMeta struct {
	filename string
	root     string

	OpenFunc FileOpener
	PathInfo *paths.Path
}

func NewFileMeta() *FileMeta {
	return &FileMeta{}
}

func (f *FileMeta) Root() string {
	return f.root
}

func (f *FileMeta) Copy() *FileMeta {
	if f == nil {
		return NewFileMeta()
	}
	c := *f
	return &c
}

func (f *FileMeta) Merge(from *FileMeta) {
	if f == nil || from == nil {
		return
	}
	dstv := reflect.Indirect(reflect.ValueOf(f))
	srcv := reflect.Indirect(reflect.ValueOf(from))

	for i := 0; i < dstv.NumField(); i++ {
		v := dstv.Field(i)
		if !v.CanSet() {
			continue
		}
		if !hreflect.IsTruthfulValue(v) {
			v.Set(srcv.Field(i))
		}
	}
}

func (f *FileMeta) NormalizedFilename() string {
	return normalizeFilename(f.FileName())
}

func (f *FileMeta) Open() (afero.File, error) {
	if f.OpenFunc == nil {
		return nil, errors.New("OpenFunc not set")
	}
	return f.OpenFunc()
}

func (f *FileMeta) FileName() string {
	return f.filename
}

func (f *FileMeta) Path() *paths.Path {
	return f.PathInfo
}

func (f *FileMeta) SetPath(path *paths.Path) {
	f.PathInfo = path
}

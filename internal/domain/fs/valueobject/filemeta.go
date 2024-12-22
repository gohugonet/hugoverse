package valueobject

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/hreflect"
	"github.com/spf13/afero"
	"reflect"
	"strings"
)

type MetaProvider interface {
	Meta() *FileMeta
}

type FileOpener func() (afero.File, error)

type FileMeta struct {
	filename string

	ComponentRoot string
	ComponentDir  string

	OpenFunc FileOpener
}

func NewFileMeta() *FileMeta {
	return &FileMeta{}
}

func (f *FileMeta) Component() string {
	return f.ComponentDir
}

func (f *FileMeta) Root() string {
	return f.ComponentRoot
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

func (f *FileMeta) RelativeFilename() (string, error) {
	if f.Root() == "" {
		return f.FileName(), nil
	}

	// 找到 f.Root() 第一次出现的位置
	rootIndex := strings.Index(f.FileName(), f.Root())
	if rootIndex == -1 {
		return "", fmt.Errorf("filename %s has no root %s", f.FileName(), f.Root())
	}

	// 截取从 f.Root() 开始的部分路径，并去掉 f.Root()
	relativePath := f.FileName()[rootIndex+len(f.Root()):]

	// 确保路径以 "/" 开头
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	return relativePath, nil
}

package valueobject

import (
	"errors"
	"github.com/gohugonet/hugoverse/pkg/hreflect"
	"github.com/spf13/afero"
	"os"
	"reflect"
)

type FileMeta struct {
	Name             string
	Filename         string
	Path             string
	PathWalk         string
	OriginalFilename string
	BaseDir          string

	SourceRoot string
	MountRoot  string
	Module     string

	Weight     int
	IsOrdered  bool
	IsSymlink  bool
	IsRootFile bool
	IsProject  bool
	Watch      bool

	Classifier ContentClass

	SkipDir bool

	Lang                       string
	TranslationBaseName        string
	TranslationBaseNameWithExt string
	Translations               []string

	Fs           afero.Fs
	OpenFunc     func() (afero.File, error)
	JoinStatFunc func(name string) (FileMetaInfo, error)

	// Include only files or directories that match.
	//InclusionFilter *glob.FilenameFilter
}

func NewFileMeta() *FileMeta {
	return &FileMeta{}
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

func (f *FileMeta) Open() (afero.File, error) {
	if f.OpenFunc == nil {
		return nil, errors.New("OpenFunc not set")
	}
	return f.OpenFunc()
}

func (f *FileMeta) JoinStat(name string) (FileMetaInfo, error) {
	if f.JoinStatFunc == nil {
		return nil, os.ErrNotExist
	}
	return f.JoinStatFunc(name)
}

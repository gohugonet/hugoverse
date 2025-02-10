package application

import (
	fsVO "github.com/mdfriday/hugoverse/internal/domain/fs/valueobject"
	"github.com/mdfriday/hugoverse/pkg/fs/static"
	"github.com/spf13/afero"
)

func staticCopy(fs map[string]*fsVO.ComponentFs, publish afero.Fs) error {
	var fss []afero.Fs
	for _, v := range fs {
		fss = append(fss, v.Fs)
	}
	return static.Copy(fss, publish)
}

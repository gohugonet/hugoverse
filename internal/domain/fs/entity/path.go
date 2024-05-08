package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/spf13/afero"
)

// PathSpec holds methods that decides how paths in URLs and files in Hugo should look like.
type PathSpec struct {
	*valueobject.BaseFs
}

func (ps *PathSpec) LayoutFs() afero.Fs {
	return ps.BaseFs.SourceFilesystems.Layouts.Fs
}

func (ps *PathSpec) ContentFs() afero.Fs {
	return ps.BaseFs.SourceFilesystems.Content.Fs
}

func (ps *PathSpec) AssetsFs() afero.Fs {
	return ps.BaseFs.SourceFilesystems.Assets.Fs
}

func (ps *PathSpec) ResourcesCacheFs() afero.Fs {
	return ps.BaseFs.SourceFilesystems.ResourcesCache
}

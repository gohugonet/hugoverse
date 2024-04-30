package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/spf13/afero"
)

var _ fs.FilesystemUnwrapper = (*filesystemsWrapper)(nil)

// NewBasePathFs creates a new BasePathFs.
func NewBasePathFs(source afero.Fs, path string) afero.Fs {
	return WrapFilesystem(afero.NewBasePathFs(source, path), source)
}

// NewReadOnlyFs creates a new ReadOnlyFs.
func NewReadOnlyFs(source afero.Fs) afero.Fs {
	return WrapFilesystem(afero.NewReadOnlyFs(source), source)
}

// WrapFilesystem is typically used to wrap a afero.BasePathFs to allow
// access to the underlying filesystem if needed.
func WrapFilesystem(container, content afero.Fs) afero.Fs {
	return filesystemsWrapper{Fs: container, content: content}
}

type filesystemsWrapper struct {
	afero.Fs
	content afero.Fs
}

func (w filesystemsWrapper) UnwrapFilesystem() afero.Fs {
	return w.content
}

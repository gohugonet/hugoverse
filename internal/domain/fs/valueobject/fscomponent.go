package valueobject

import "github.com/spf13/afero"

type ComponentFs struct {
	afero.Fs

	opts ComponentFsOptions
}

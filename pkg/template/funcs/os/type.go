package os

import "github.com/spf13/afero"

type Os interface {
	CheckAllowedGetEnv(name string) error
	WorkingDir() string

	WorkFs() afero.Fs
	Working() afero.Fs
	ContentFs() afero.Fs
	RelPathify(filename string, workingDir string) string
}

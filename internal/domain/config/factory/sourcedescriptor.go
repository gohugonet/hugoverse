package factory

import "github.com/spf13/afero"

// ConfigSourceDescriptor describes where to find the config (e.g. config.toml etc.).
type sourceDescriptor struct {
	fs afero.Fs

	// Path to the config file to use, e.g. /my/project/config.toml
	filename string
}

func (sd *sourceDescriptor) Fs() afero.Fs {
	return sd.fs
}

func (sd *sourceDescriptor) Filename() string {
	return sd.filename
}

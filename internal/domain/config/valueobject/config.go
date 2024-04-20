package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"github.com/gohugonet/hugoverse/pkg/herrors"
	"github.com/gohugonet/hugoverse/pkg/parser/metadecoders"
	"github.com/spf13/afero"
	"path/filepath"
)

func CheckConfigFilename(dir string, fs afero.Fs) (string, bool) {
	var (
		configFilename string
		hasConfigFile  bool
	)

LOOP:
	for _, configBaseName := range config.DefaultConfigNames {
		for _, configFormats := range config.ValidConfigFileExtensions {
			configFilename = filepath.Join(dir, fmt.Sprintf("%s.%s", configBaseName, configFormats))
			hasConfigFile, _ = afero.Exists(fs, configFilename)
			if hasConfigFile {
				break LOOP
			}
		}
	}

	return configFilename, hasConfigFile
}

// FromFile loads the configuration from the given filename.
func FromFile(fs afero.Fs, filename string) (config.Provider, error) {
	m, err := LoadConfigFromFile(fs, filename)
	if err != nil {
		fe := herrors.UnwrapFileError(err)
		if fe != nil {
			pos := fe.Position()
			pos.Filename = filename
			_ = fe.UpdatePosition(pos)
			return nil, err
		}
		return nil, herrors.NewFileErrorFromFile(err, filename, fs, nil)
	}
	return NewFrom(m), nil
}

func LoadConfigFromFile(fs afero.Fs, filename string) (map[string]any, error) {
	m, err := metadecoders.Default.UnmarshalFileToMap(fs, filename)
	if err != nil {
		return nil, err
	}
	return m, nil
}

package testkit

import "github.com/spf13/afero"

func MkTestTempDir(fs afero.Fs, prefix string) (string, func(), error) {
	tempDir, err := afero.TempDir(fs, "", prefix)
	if err != nil {
		return "", nil, err
	}

	return tempDir, func() { fs.RemoveAll(tempDir) }, nil
}

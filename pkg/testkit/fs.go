package testkit

import (
	"bytes"
	"fmt"
	"github.com/spf13/afero"
	"golang.org/x/tools/txtar"
	"path/filepath"
)

func prepareFS(workingDir string, files string) {
	data := txtar.Parse([]byte(files))
	for _, f := range data.Files {
		filename := filepath.Join(workingDir, f.Name)
		data := bytes.TrimSuffix(f.Data, []byte("\n"))

		err := testOs.MkdirAll(filepath.Dir(filename), 0777)
		if err != nil {
			fmt.Println(err)
		}
		err = afero.WriteFile(testOs, filename, data, 0666)
		if err != nil {
			fmt.Println(err)
		}
	}
}

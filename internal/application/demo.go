package application

import (
	"bytes"
	"fmt"
	"github.com/spf13/afero"
	"golang.org/x/tools/txtar"
	"path/filepath"
)

func NewDemo() (string, error) {
	var demoOs = &afero.OsFs{}
	tempDir, clean, err := CreateTempDir(demoOs, "hugoverse-temp-dir")
	if err != nil {
		clean()
		return "", err
	}

	var afs afero.Fs
	afs = afero.NewOsFs()
	prepareFS(tempDir, afs)

	return tempDir, nil
}

// CreateTempDir creates a temp dir in the given filesystem and
// returns the dirnam and a func that removes it when done.
func CreateTempDir(fs afero.Fs, prefix string) (string, func(), error) {
	tempDir, err := afero.TempDir(fs, "", prefix)
	if err != nil {
		return "", nil, err
	}

	return tempDir, func() { fs.RemoveAll(tempDir) }, nil
}

func prepareFS(workingDir string, afs afero.Fs) {
	files := `
-- config.toml --
theme = "mytheme"
contentDir = "mycontent"
-- myproject.txt --
Hello project!
-- themes/mytheme/mytheme.txt --
Hello theme!
-- mycontent/blog/post.md --
### first blog
Hello Blog
-- layouts/index.html --
<p><!-- HTML comment -->abc</p>
{{.Content}}
-- layouts/_default/single.html --
<p>hello single page</p>
{{.Content}}`
	data := txtar.Parse([]byte(files))
	for _, f := range data.Files {
		filename := filepath.Join(workingDir, f.Name)
		data := bytes.TrimSuffix(f.Data, []byte("\n"))

		err := afs.MkdirAll(filepath.Dir(filename), 0777)
		if err != nil {
			fmt.Println(err)
		}
		err = afero.WriteFile(afs, filename, data, 0666)
		if err != nil {
			fmt.Println(err)
		}
	}
}

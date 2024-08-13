package testkit

import "fmt"

func MkTestResource() (string, func(), error) {
	tempDir, clean, err := MkTestTempDir(testOs, "go-hugoverse-temp-dir")
	if err != nil {
		return "", clean, err
	}

	files := fmt.Sprintf(`
-- config.toml --
%s
-- go.mod --
%s
-- assets/book.scss --
%s
-- assets/_defaults.scss --
%s

`, configEmptyContent, goModEmptyContent, sassBook, sassDefault)

	prepareFS(tempDir, files)
	return tempDir, clean, nil
}

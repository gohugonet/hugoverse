package testkit

import "fmt"

func MkTestModule() (string, func(), error) {
	tempDir, clean, err := MkTestTempDir(testOs, "go-hugoverse-temp-dir")
	if err != nil {
		return "", clean, err
	}

	files := fmt.Sprintf(`
-- config.toml --
%s
-- go.mod --
%s
`, configContent, goModContent)

	prepareFS(tempDir, files)
	return tempDir, clean, nil
}

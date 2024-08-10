package testkit

import "fmt"

func MkTestContent() (string, func(), error) {
	tempDir, clean, err := MkTestTempDir(testOs, "go-hugoverse-temp-dir")
	if err != nil {
		return "", clean, err
	}

	files := fmt.Sprintf(`
-- config.toml --
%s
-- go.mod --
%s
-- content/_index.md --
%s
-- content/post/index.md --
%s
-- content.en/_index.md --
%s
-- content.en/post/index.md --
%s
-- blog/_index.md --
%s
-- blog/abc.md --
%s
`, configContent, goModContent, post1Content, post2Content, post1Content, post2Content, blog1Content, blog2Content)

	prepareFS(tempDir, files)
	return tempDir, clean, nil
}

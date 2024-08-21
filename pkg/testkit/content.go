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
-- content/blog/_index.md --
%s
-- content/blog/abc.md --
%s
-- content/post/ddd/_index.md --
%s
`, configContent, goModContent, post1Content, post2Content, post1Content, post2Content, blog1Content, blog2Content, post1Content)

	prepareFS(tempDir, files)
	return tempDir, clean, nil
}

func MkTestContentHub() (string, func(), error) {
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
-- content/blog/_index.md --
%s
-- content/blog/abc.md --
%s
-- content/post/ddd/_index.md --
%s
`, configMulLangContent, goModContent, post1Content, post2Content, post1Content, post2Content, blog1Content, blog2Content, post1Content)

	prepareFS(tempDir, files)
	return tempDir, clean, nil
}

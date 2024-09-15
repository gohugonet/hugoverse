package testkit

import "fmt"

func MkTestSite() (string, func(), error) {
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
-- content/docs/_index.md --
%s
-- content/docs/api/_index.md --
%s
-- content/docs/api/book/index.md --
%s
`, configContent, goModContent,
		post1Content, post2Content, post1Content, post2Content, blog1Content, blog2Content, post1Content,
		post1Content, post2Content, post1Content)

	prepareFS(tempDir, files)
	return tempDir, clean, nil
}

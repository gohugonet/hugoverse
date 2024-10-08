package testkit

import (
	"fmt"
)

func MkTestConfig() (string, func(), error) {
	tempDir, clean, err := MkTestTempDir(testOs, "go-hugoverse-temp-dir")
	if err != nil {
		return "", clean, err
	}

	files := fmt.Sprintf(`
-- config.toml --
%s
`, configContent)

	prepareFS(tempDir, files)
	return tempDir, clean, nil
}

const configContent = `
baseURL = '/'
title = 'Deep Dive into Hugo'
theme = "github.com/sunwei/hugo-book"
defaultContentLanguage = "zh"
googleAnalytics = "G-STPKPBQR5Y"

# Book configuration
disablePathToLower = true
enableGitInfo = true

# Needed for mermaid/katex shortcodes
[markup]
[markup.goldmark.renderer]
  unsafe = true

[markup.tableOfContents]
  startLevel = 1

[module]
[[module.imports]]
path = 'github.com/sunwei/hugo-book'

[languages]
[languages.zh]
  languageName = 'Chinese'
  contentDir = 'content'
  weight = 1

[languages.en]
  languageName = 'English'
  contentDir = 'content.en'
  weight = 2

[menu]
# [[menu.before]]
[[menu.after]]
  name = "Github"
  url = "https://github.com/sunwei/hugo-book"
  weight = 10

[[menu.after]]
  name = "Hugo Themes"
  url = "https://themes.gohugo.io/hugo-book/"
  weight = 20

[params]
  # (Optional, default light) Sets color theme: light, dark or auto.
  # Theme 'auto' switches between dark and light modes based on browser/os preferences
  BookTheme = 'light'

  # (Optional, default true) Controls table of contents visibility on right side of pages.
  # Start and end levels can be controlled with markup.tableOfContents setting.
  # You can also specify this parameter per page in front matter.
  BookToC = true

  # (Optional, default none) Set the path to a logo for the book. If the logo is
  # /static/logo.png then the path would be logo.png
  # BookLogo = 'logo.png'

  # Set source repository location.
  # Used for 'Last Modified' and 'Edit this page' links.
  BookRepo = 'https://github.com/sunwei/hugo-notes'

  # (Optional, default 'commit') Specifies commit portion of the link to the page's last modified
  # commit hash for 'doc' page type.
  # Requires 'BookRepo' param.
  # Value used to construct a URL consisting of BookRepo/BookCommitPath/<commit-hash>
  # Github uses 'commit', Bitbucket uses 'commits'
  # BookCommitPath = 'commit'

  # Enable "Edit this page" links for 'doc' page type.
  # Disabled by default. Uncomment to enable. Requires 'BookRepo' param.
  # Edit path must point to root directory of repo.
  BookEditPath = 'edit/main'

  # Configure the date format used on the pages
  # - In git information
  # - In blog posts
  BookDateFormat = 'January 2, 2006'

  # (Optional, default true) Enables search function with flexsearch,
  # Index is built on fly, therefore it might slowdown your website.
  # Configuration for indexing can be adjusted in i18n folder per language.
  BookSearch = true

  # (Optional, default true) Enables comments template on pages
  # By default partals/docs/comments.html includes Disqus template
  # See https://gohugo.io/content-management/comments/#configure-disqus
  # Can be overwritten by same param in page frontmatter
  BookComments = true

  # /!\ This is an experimental feature, might be removed or changed at any time
  # (Optional, experimental, default false) Enables portable links and link checks in markdown pages.
  # Portable links meant to work with text editors and let you write markdown without {{< relref >}} shortcode
  # Theme will print warning if page referenced in markdown does not exists.
  BookPortableLinks = true

  # /!\ This is an experimental feature, might be removed or changed at any time
  # (Optional, experimental, default false) Enables service worker that caches visited pages and resources for offline use.
  BookServiceWorker = true

  # /!\ This is an experimental feature, might be removed or changed at any time
  # (Optional, experimental, default false) Enables a drop-down menu for translations only if a translation is present.
  BookTranslatedOnly = false

`

const configEmptyContent = `
title = 'Deep Dive into Hugo: Becoming an Expert in the Static Site Generator Domain'

`

const configMulLangContent = `
title = 'Deep Dive into Hugo: Becoming an Expert in the Static Site Generator Domain'

defaultContentLanguage = "zh"

[languages]
[languages.zh]
  languageName = 'Chinese'
  contentDir = 'content'
  weight = 1

[languages.en]
  languageName = 'English'
  contentDir = 'content.en'
  weight = 2

`

package valueobject

import (
	"fmt"
	godartsassv1 "github.com/bep/godartsass"
	"github.com/bep/godartsass/v2"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/types/css"
	"github.com/spf13/afero"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type ImportResolverV1 struct {
	godartsass.ImportResolver
}

func (t ImportResolverV1) Load(url string) (godartsassv1.Import, error) {
	res, err := t.ImportResolver.Load(url)
	return godartsassv1.Import{Content: res.Content, SourceSyntax: godartsassv1.SourceSyntax(res.SourceSyntax)}, err
}

const (
	HugoVarsNamespace = "hugo:vars"
)

type ImportResolver struct {
	BaseDir           string
	FsService         resources.Fs
	DependencyManager identity.Manager
	VarsStylesheet    godartsass.Import
}

func (t ImportResolver) CanonicalizeURL(url string) (string, error) {
	if url == HugoVarsNamespace {
		return url, nil
	}

	filePath, isURL := paths.UrlToFilename(url)
	var prevDir string
	var pathDir string
	if isURL {
		var found bool
		prevDir, found = t.FsService.AssetsFsMakePathRelative(filepath.Dir(filePath), true)

		if !found {
			// Not a member of this filesystem, let Dart Sass handle it.
			return "", nil
		}
	} else {
		prevDir = t.BaseDir
		pathDir = path.Dir(url)
	}

	basePath := filepath.Join(prevDir, pathDir)
	name := filepath.Base(filePath)

	// Pick the first match.
	var namePatterns []string
	if strings.Contains(name, ".") {
		namePatterns = []string{"_%s", "%s"}
	} else if strings.HasPrefix(name, "_") {
		namePatterns = []string{"_%s.scss", "_%s.sass", "_%s.css"}
	} else {
		namePatterns = []string{"_%s.scss", "%s.scss", "_%s.sass", "%s.sass", "_%s.css", "%s.css"}
	}

	name = strings.TrimPrefix(name, "_")

	for _, namePattern := range namePatterns {
		filenameToCheck := filepath.Join(basePath, fmt.Sprintf(namePattern, name))
		fi, err := t.FsService.AssetsFs().Stat(filenameToCheck)
		if err == nil {
			if fim, ok := fi.(fs.FileMetaInfo); ok {
				t.DependencyManager.AddIdentity(identity.CleanStringIdentity(filenameToCheck))

				return "file:///" + filepath.ToSlash(fim.FileName()), nil
			}
		}
	}

	// Not found, let Dart Sass handle it
	return "", nil
}

func (t ImportResolver) Load(url string) (godartsass.Import, error) {
	if url == HugoVarsNamespace {
		return t.VarsStylesheet, nil
	}
	filename, _ := paths.UrlToFilename(url)
	b, err := afero.ReadFile(t.FsService.AssetsFs(), filename)

	sourceSyntax := godartsass.SourceSyntaxSCSS
	if strings.HasSuffix(filename, ".sass") {
		sourceSyntax = godartsass.SourceSyntaxSASS
	} else if strings.HasSuffix(filename, ".css") {
		sourceSyntax = godartsass.SourceSyntaxCSS
	}

	return godartsass.Import{Content: string(b), SourceSyntax: sourceSyntax}, err
}

func CreateVarsStyleSheet(vars map[string]any) string {
	if vars == nil {
		return ""
	}
	var varsStylesheet string

	var varsSlice []string
	for k, v := range vars {
		var prefix string
		if !strings.HasPrefix(k, "$") {
			prefix = "$"
		}

		switch v.(type) {
		case css.QuotedString:
			// Marked by the user as a string that needs to be quoted.
			varsSlice = append(varsSlice, fmt.Sprintf("%s%s: %q;", prefix, k, v))
		default:
			if isTypedCSSValue(v) {
				// E.g. 24px, 1.5rem, 10%, hsl(0, 0%, 100%), calc(24px + 36px), #fff, #ffffff.
				varsSlice = append(varsSlice, fmt.Sprintf("%s%s: %v;", prefix, k, v))
			} else {
				// unquote will preserve quotes around URLs etc. if needed.
				varsSlice = append(varsSlice, fmt.Sprintf("%s%s: unquote(%q);", prefix, k, v))
			}
		}
	}
	sort.Strings(varsSlice)
	varsStylesheet = strings.Join(varsSlice, "\n")
	return varsStylesheet
}

// isTypedCSSValue returns true if the given string is a CSS value that
// we should preserve the type of, as in: Not wrap it in quotes.
func isTypedCSSValue(v any) bool {
	switch s := v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, css.UnquotedString:
		return true
	case string:
		if isCSSColor.MatchString(s) {
			return true
		}
		if isCSSFunc.MatchString(s) {
			return true
		}
		if isCSSUnit.MatchString(s) {
			return true
		}

	}

	return false
}

var (
	isCSSColor = regexp.MustCompile(`^#[0-9a-fA-F]{3,6}$`)
	isCSSFunc  = regexp.MustCompile(`^([a-zA-Z-]+)\(`)
	isCSSUnit  = regexp.MustCompile(`^([0-9]+)(\.[0-9]+)?([a-zA-Z-%]+)$`)
)

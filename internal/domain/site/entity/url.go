package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/site/valueobject"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/text"
	"net/url"
	"path"
	"strings"
)

type URL struct {
	Base      string
	Canonical bool

	*valueobject.BaseURL
}

func (u *URL) setup() error {
	bu, err := valueobject.NewBaseURLFromString(u.Base)
	if err != nil {
		return err
	}
	u.BaseURL = &bu
	return nil
}

func (u *URL) IsAbsURL(in string) (bool, error) {
	// Fast path.
	if strings.HasPrefix(in, "http://") || strings.HasPrefix(in, "https://") {
		return true, nil
	}
	pu, err := url.Parse(in)
	if err != nil {
		return false, err
	}
	return pu.IsAbs(), nil
}

func (u *URL) startWithBaseUrlRoot(in string) bool {
	return strings.HasPrefix(in, u.GetRoot(in))
}

func (u *URL) isProtocolRelPath(in string) bool {
	return strings.HasPrefix(in, "//")
}

func (u *URL) trimBaseUrlRoot(in string) string {
	root := u.GetRoot(in)
	if strings.HasSuffix(in, root) {
		return strings.TrimSuffix(in, root)
	}
	return in
}

func (u *URL) addContextRoot(in string) string {
	out := in
	if !u.Canonical {
		out = paths.AddContextRoot(u.GetRoot(in), in)
	}
	return out
}

func (u *URL) handleRootSuffix(in, url string) string {
	if in == "" && strings.HasSuffix(u.BaseURL.GetRoot(in), "/") {
		url += "/"
	}
	return url
}

func (u *URL) handlePrefix(url string) string {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	return url
}

func (s *Site) RelURL(in string) string {
	isAbs, err := s.URL.IsAbsURL(in)
	if err != nil {
		return in
	}

	if (!s.URL.startWithBaseUrlRoot(in) && isAbs) || s.URL.isProtocolRelPath(in) {
		return in
	}

	u := s.URL.trimBaseUrlRoot(in)

	// For resources only, no need to add language prefix

	u = s.URL.addContextRoot(u)
	u = s.URL.handleRootSuffix(in, u)

	//u = s.URL.handlePrefix(u) // our use case include preview, just put it relative to the html file

	return u
}

func (s *Site) AbsURL(in string) string {
	isAbs, err := s.IsAbsURL(in)
	if err != nil {
		return in
	}
	if isAbs || s.URL.isProtocolRelPath(in) {
		// It  is already  absolute, return it as is.
		return in
	}

	baseURL := s.URL.GetRoot(in)

	if s.isMultipleLanguage() {
		prefix := s.LanguagePrefix()
		if prefix != "" {
			hasPrefix := false
			// avoid adding language prefix if already present
			in2 := in
			if strings.HasPrefix(in, "/") {
				in2 = in[1:]
			}
			if in2 == prefix {
				hasPrefix = true
			} else {
				hasPrefix = strings.HasPrefix(in2, prefix+"/")
			}

			if !hasPrefix {
				addSlash := in == "" || strings.HasSuffix(in, "/")
				in = path.Join(prefix, in)

				if addSlash {
					in += "/"
				}
			}
		}
	}

	return paths.MakePermalink(baseURL, in).String()
}

// URLize is similar to MakePath, but with Unicode handling
// Example:
//
//	uri: Vim (text editor)
//	urlize: vim-text-editor
func (u *URL) URLize(uri string) string {
	return u.URLEscape(u.MakePathSanitized(uri))
}

// MakePathSanitized creates a Unicode-sanitized string, with the spaces replaced
func (u *URL) MakePathSanitized(s string) string {
	return strings.ToLower(u.MakePath(s))
}

// URLEscape escapes unicode letters.
func (u *URL) URLEscape(uri string) string {
	// escape unicode letters
	parsedURI, err := url.Parse(uri)
	if err != nil {
		// if net/url can not parse URL it means Sanitize works incorrectly
		panic(err)
	}
	x := parsedURI.String()
	return x
}

// MakePath takes a string with any characters and replace it
// so the string could be used in a path.
// It does so by creating a Unicode-sanitized string, with the spaces replaced,
// whilst preserving the original casing of the string.
// E.g. Social Media -> Social-Media
func (u *URL) MakePath(s string) string {
	s = paths.Sanitize(s)
	s = text.RemoveAccentsString(s)
	return s
}

func (u *URL) BasePathNoSlash() string {
	return u.BaseURL.BasePathNoTrailingSlash
}

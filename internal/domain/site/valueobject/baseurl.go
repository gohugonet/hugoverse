package valueobject

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// A BaseURL in Hugo is normally on the form scheme://path, but the
// form scheme: is also valid (mailto:hugo@rules.com).
type BaseURL struct {
	url                     *url.URL
	WithPath                string
	WithPathNoTrailingSlash string
	WithoutPath             string
	BasePath                string
	BasePathNoTrailingSlash string
}

func (b BaseURL) GetRoot(path string) string {
	if strings.HasPrefix(path, "/") {
		// Treat it as relative to the server root.
		return b.WithoutPath
	}
	// Treat it as relative to the baseURL.
	return b.WithPath
}

func (b BaseURL) String() string {
	return b.WithPath
}

func (b BaseURL) Path() string {
	return b.url.Path
}

func (b BaseURL) Port() int {
	p, _ := strconv.Atoi(b.url.Port())
	return p
}

// HostURL returns the URL to the host root without any path elements.
func (b BaseURL) HostURL() string {
	return strings.TrimSuffix(b.String(), b.Path())
}

// WithProtocol returns the BaseURL prefixed with the given protocol.
// The Protocol is normally of the form "scheme://", i.e. "webcal://".
func (b BaseURL) WithProtocol(protocol string) (BaseURL, error) {
	u := b.URL()

	scheme := protocol
	isFullProtocol := strings.HasSuffix(scheme, "://")
	isOpaqueProtocol := strings.HasSuffix(scheme, ":")

	if isFullProtocol {
		scheme = strings.TrimSuffix(scheme, "://")
	} else if isOpaqueProtocol {
		scheme = strings.TrimSuffix(scheme, ":")
	}

	u.Scheme = scheme

	if isFullProtocol && u.Opaque != "" {
		u.Opaque = "//" + u.Opaque
	} else if isOpaqueProtocol && u.Opaque == "" {
		return BaseURL{}, fmt.Errorf("cannot determine BaseURL for protocol %q", protocol)
	}

	return newBaseURLFromURL(u)
}

func (b BaseURL) WithPort(port int) (BaseURL, error) {
	u := b.URL()
	u.Host = u.Hostname() + ":" + strconv.Itoa(port)
	return newBaseURLFromURL(u)
}

// URL returns a copy of the internal URL.
// The copy can be safely used and modified.
func (b BaseURL) URL() *url.URL {
	c := *b.url
	return &c
}

func NewBaseURLFromString(b string) (BaseURL, error) {
	u, err := url.Parse(b)
	if err != nil {
		return BaseURL{}, err
	}
	return newBaseURLFromURL(u)
}

func newBaseURLFromURL(u *url.URL) (BaseURL, error) {
	// A baseURL should always have a trailing slash, see #11669.
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	baseURL := BaseURL{url: u, WithPath: u.String(), WithPathNoTrailingSlash: strings.TrimSuffix(u.String(), "/")}
	baseURLNoPath := baseURL.URL()
	baseURLNoPath.Path = ""
	baseURL.WithoutPath = baseURLNoPath.String()
	baseURL.BasePath = u.Path
	baseURL.BasePathNoTrailingSlash = strings.TrimSuffix(u.Path, "/")

	return baseURL, nil
}

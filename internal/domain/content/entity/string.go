package entity

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"regexp"
	"strings"
	"unicode"
)

var rxList map[*regexp.Regexp][]byte

func init() {
	// Compile regex once to use in stringToSlug().
	// We store the compiled regex as the key
	// and assign the replacement as the map's value.
	rxList = map[*regexp.Regexp][]byte{
		regexp.MustCompile("`[-]+`"):                  []byte("-"),
		regexp.MustCompile("[[:space:]]"):             []byte("-"),
		regexp.MustCompile("[[:blank:]]"):             []byte(""),
		regexp.MustCompile("`[^a-z0-9]`i"):            []byte("-"),
		regexp.MustCompile("[!/:-@[-`{-~]"):           []byte(""),
		regexp.MustCompile("/[^\x20-\x7F]/"):          []byte(""),
		regexp.MustCompile("`&(amp;)?#?[a-z0-9]+;`i"): []byte("-"),
		regexp.MustCompile("`&([a-z])(acute|uml|circ|grave|ring|cedil|slash|tilde|caron|lig|quot|rsquo);`i"): []byte("\\1"),
	}
}

// modified version of: https://www.socketloop.com/tutorials/golang-format-strings-to-seo-friendly-url-example
func stringToSlug(s string) (string, error) {
	src := []byte(strings.ToLower(s))

	// Range over compiled regex and replacements from init().
	for rx := range rxList {
		src = rx.ReplaceAll(src, rxList[rx])
	}

	str := strings.Replace(string(src), "'", "", -1)
	str = strings.Replace(str, `"`, "", -1)
	str = strings.Replace(str, "&", "-", -1)

	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	slug, _, err := transform.String(t, str)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(slug), nil
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

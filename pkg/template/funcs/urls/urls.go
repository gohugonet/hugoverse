// Copyright 2017 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package urls provides template functions to deal with URLs.
package urls

import (
	"fmt"
	"net/url"

	"github.com/spf13/cast"
)

// New returns a new instance of the urls-namespaced template functions.
func New(u URL, r RefSource) *Namespace {
	return &Namespace{
		url: u,
		ref: r,
	}
}

// Namespace provides template functions for the "urls" namespace.
type Namespace struct {
	url URL
	ref RefSource
}

// AbsURL takes the string s and converts it to an absolute URL.
func (ns *Namespace) AbsURL(s any) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", nil
	}

	return ns.url.AbsURL(ss), nil
}

// Parse parses rawurl into a URL structure. The rawurl may be relative or
// absolute.
func (ns *Namespace) Parse(rawurl any) (*url.URL, error) {
	s, err := cast.ToStringE(rawurl)
	if err != nil {
		return nil, fmt.Errorf("error in Parse: %w", err)
	}

	return url.Parse(s)
}

// RelURL takes the string s and prepends the relative path according to a
// page's position in the project directory structure.
func (ns *Namespace) RelURL(s any) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", nil
	}

	return ns.url.RelURL(ss), nil
}

// URLize returns the the strings s formatted as an URL.
func (ns *Namespace) URLize(s any) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", nil
	}
	return ns.url.URLize(ss), nil
}

// Ref returns the absolute URL path to a given content item from Page p.
func (ns *Namespace) Ref(p any, args any) (string, error) {
	argsm, err := ns.refArgsToMap(args)
	if err != nil {
		return "", err
	}
	s, err := ns.ref.RelRefFrom(argsm, p)
	return s, err
}

// RelRef returns the relative URL path to a given content item from Page p.
func (ns *Namespace) RelRef(p any, args any) (string, error) {
	argsm, err := ns.refArgsToMap(args)
	if err != nil {
		return "", err
	}

	s, err := ns.ref.RelRefFrom(argsm, p)
	return s, err
}

func (ns *Namespace) refArgsToMap(args any) (map[string]any, error) {
	var (
		s  string
		of string
	)

	v := args
	if _, ok := v.([]any); ok {
		v = cast.ToStringSlice(v)
	}

	switch v := v.(type) {
	case map[string]any:
		return v, nil
	case map[string]string:
		m := make(map[string]any)
		for k, v := range v {
			m[k] = v
		}
		return m, nil
	case []string:
		if len(v) == 0 || len(v) > 2 {
			return nil, fmt.Errorf("invalid number of arguments to ref")
		}
		// These were the options before we introduced the map type:
		s = v[0]
		if len(v) == 2 {
			of = v[1]
		}
	default:
		var err error
		s, err = cast.ToStringE(args)
		if err != nil {
			return nil, err
		}

	}

	return map[string]any{
		"path":         s,
		"outputFormat": of,
	}, nil
}

// RelLangURL takes the string s and prepends the relative path according to a
// page's position in the project directory structure and the current language.
func (ns *Namespace) RelLangURL(s any) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	return ns.url.RelURL(ss), nil
}

// AbsLangURL the string s and converts it to an absolute URL according
// to a page's position in the project directory structure and the current
// language.
func (ns *Namespace) AbsLangURL(s any) (string, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return "", err
	}

	return ns.url.AbsURL(ss), nil
}

// JoinPath joins the provided elements into a URL string and cleans the result
// of any ./ or ../ elements. If the argument list is empty, JoinPath returns
// an empty string.
func (ns *Namespace) JoinPath(elements ...any) (string, error) {
	if len(elements) == 0 {
		return "", nil
	}

	var selements []string
	for _, e := range elements {
		switch v := e.(type) {
		case []string:
			selements = append(selements, v...)
		case []any:
			for _, e := range v {
				se, err := cast.ToStringE(e)
				if err != nil {
					return "", err
				}
				selements = append(selements, se)
			}
		default:
			se, err := cast.ToStringE(e)
			if err != nil {
				return "", err
			}
			selements = append(selements, se)
		}
	}

	result, err := url.JoinPath(selements[0], selements[1:]...)
	if err != nil {
		return "", err
	}
	return result, nil
}

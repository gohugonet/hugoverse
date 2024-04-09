package valueobject

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	baseFileBase = "baseof"
)

type BaseOf struct {
	baseof      map[string]TemplateInfo
	needsBaseof map[string]TemplateInfo
}

func NewBaseOf() *BaseOf {
	return &BaseOf{
		baseof:      make(map[string]TemplateInfo),
		needsBaseof: make(map[string]TemplateInfo),
	}
}

func (bo *BaseOf) GetBaseOf(key string) (TemplateInfo, bool) {
	info, ok := bo.baseof[key]
	return info, ok
}

func (bo *BaseOf) GetNeedsBaseOf(key string) (TemplateInfo, bool) {
	info, ok := bo.needsBaseof[key]
	return info, ok
}

func (bo *BaseOf) AddBaseOf(key string, info TemplateInfo) {
	bo.baseof[key] = info
}

func (bo *BaseOf) AddNeedsBaseOf(key string, info TemplateInfo) {
	bo.needsBaseof[key] = info
}

func (bo *BaseOf) IsBaseTemplatePath(path string) bool {
	return strings.Contains(filepath.Base(path), baseFileBase)
}

func (bo *BaseOf) NeedsBaseOf(name, rawContent string) bool {
	return !bo.noBaseNeeded(name) && bo.needsBaseTemplate(rawContent)
}

func (bo *BaseOf) noBaseNeeded(name string) bool {
	if strings.HasPrefix(name, "shortcodes/") || strings.HasPrefix(name, "partials/") {
		return true
	}
	return strings.Contains(name, "_markup/")
}

var baseTemplateDefineRe = regexp.MustCompile(`^{{-?\s*define`)

// needsBaseTemplate returns true if the first non-comment template block is a
// define block.
// If a base template does not exist, we will handle that when it's used.
func (bo *BaseOf) needsBaseTemplate(templ string) bool {
	idx := -1
	inComment := false
	for i := 0; i < len(templ); {
		if !inComment && strings.HasPrefix(templ[i:], "{{/*") {
			inComment = true
			i += 4
		} else if !inComment && strings.HasPrefix(templ[i:], "{{- /*") {
			inComment = true
			i += 6
		} else if inComment && strings.HasPrefix(templ[i:], "*/}}") {
			inComment = false
			i += 4
		} else if inComment && strings.HasPrefix(templ[i:], "*/ -}}") {
			inComment = false
			i += 6
		} else {
			r, size := utf8.DecodeRuneInString(templ[i:])
			if !inComment {
				if strings.HasPrefix(templ[i:], "{{") {
					idx = i
					break
				} else if !unicode.IsSpace(r) {
					break
				}
			}
			i += size
		}
	}

	if idx == -1 {
		return false
	}

	return baseTemplateDefineRe.MatchString(templ[idx:])
}

package entity

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/content"
	"strings"
)

func GetNamespace(contentType, status string) string {
	ns := contentType
	if !(status == "" || status == string(content.Public)) {
		ns = fmt.Sprintf("%s__%s", contentType, status)
	}
	return ns
}

func isPublicNamespace(ns string) bool {
	return !strings.Contains(ns, "__")
}

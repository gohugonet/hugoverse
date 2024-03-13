package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content"
)

func GetNamespace(contentType, status string) string {
	ns := contentType
	if !(status == "" || status == string(content.Public)) {
		ns = fmt.Sprintf("%s__%s", contentType, status)
	}
	return ns
}

package entity

import (
	"encoding/base64"
	"fmt"
	"time"
)

// newEtag generates a new Etag for response caching
func newEtag() string {
	now := fmt.Sprintf("%d", time.Now().Unix())
	etag := base64.StdEncoding.EncodeToString([]byte(now))

	return etag
}

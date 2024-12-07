package valueobject

import (
	"crypto/sha256"
	"encoding/hex"
)

func Hash(fields []string) string {
	data := ""
	for _, field := range fields {
		data += field
	}

	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}

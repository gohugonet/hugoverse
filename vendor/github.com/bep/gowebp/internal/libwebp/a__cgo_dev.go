//go:build dev
// +build dev

package libwebp

// #cgo LDFLAGS: -lwebp
// #cgo CFLAGS: -DLIBWEBP_NO_SRC
import "C"

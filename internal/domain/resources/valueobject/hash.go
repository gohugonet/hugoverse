package valueobject

import (
	"fmt"
	"github.com/mdfriday/hugoverse/pkg/helpers"
	"github.com/mdfriday/hugoverse/pkg/io"
	"sync"
)

type ResourceHash struct {
	Value    string
	Size     int64
	InitOnce sync.Once
}

func (r *ResourceHash) Setup(l io.ReadSeekCloserProvider) error {
	var initErr error
	r.InitOnce.Do(func() {
		var hash string
		var size int64
		f, err := l.ReadSeekCloser()
		if err != nil {
			initErr = fmt.Errorf("failed to open source: %w", err)
			return
		}
		defer f.Close()
		hash, size, err = helpers.MD5FromReaderFast(f)
		if err != nil {
			initErr = fmt.Errorf("failed to calculate hash: %w", err)
			return
		}
		r.Value = hash
		r.Size = size
	})

	return initErr
}

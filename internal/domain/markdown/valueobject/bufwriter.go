package valueobject

import (
	"bytes"
	"math/bits"
)

type BufWriter struct {
	*bytes.Buffer
}

const maxInt = 1<<(bits.UintSize-1) - 1

func (b *BufWriter) Available() int {
	return maxInt
}

func (b *BufWriter) Buffered() int {
	return b.Len()
}

func (b *BufWriter) Flush() error {
	return nil
}

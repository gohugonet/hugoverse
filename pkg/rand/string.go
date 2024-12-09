package rand

import (
	"golang.org/x/exp/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func ShortString(length int) string {
	// 使用时间作为种子，确保随机性
	rand.Seed(uint64(time.Now().UnixNano()))

	// 生成随机字符串
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

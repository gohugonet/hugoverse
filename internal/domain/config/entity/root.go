package entity

import (
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/config/valueobject"
	"strconv"
	"time"
)

type Root struct {
	valueobject.RootConfig
}

func (r Root) DefaultTheme() string {
	if len(r.Theme) > 0 {
		return r.Theme[0]
	}
	return ""
}

func (r Root) CompiledTimeout() (time.Duration, error) {
	s := r.Timeout
	if _, err := strconv.Atoi(s); err == nil {
		// A number, assume seconds.
		s = s + "s"
	}
	timeout, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse timeout: %s", err)
	}

	return timeout, nil
}

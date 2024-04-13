package valueobject

import (
	"fmt"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"strconv"
	"time"
)

// ConfigCompiled holds values and functions that are derived from the config.
type ConfigCompiled struct {
	Timeout time.Duration
}

type BaseConfig struct {
	RootConfig

	// For internal use only.
	C *ConfigCompiled `mapstructure:"-" json:"-"`

	// Module configuration.
	Module ModuleConfig `mapstructure:"-"`

	// The languages configuration sections maps a language code (a string) to a configuration object for that language.
	Languages map[string]LanguageConfig `mapstructure:"-"`
}

func (c *BaseConfig) CompileConfig(logger loggers.Logger) error {
	s := c.Timeout
	if _, err := strconv.Atoi(s); err == nil {
		// A number, assume seconds.
		s = s + "s"
	}
	timeout, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("failed to parse timeout: %s", err)
	}

	c.C = &ConfigCompiled{
		Timeout: timeout,
	}

	for _, s := range AllDecoderSetups {
		if getCompiler := s.GetCompiler; getCompiler != nil {
			if err := getCompiler(c).CompileConfig(logger); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *BaseConfig) CloneForLang() *BaseConfig {
	x := c
	x.C = nil
	copyStringSlice := func(in []string) []string {
		if in == nil {
			return nil
		}
		out := make([]string, len(in))
		copy(out, in)
		return out
	}

	// Copy all the slices to avoid sharing.
	x.DisableKinds = copyStringSlice(x.DisableKinds)
	x.DisableLanguages = copyStringSlice(x.DisableLanguages)
	x.MainSections = copyStringSlice(x.MainSections)
	x.IgnoreLogs = copyStringSlice(x.IgnoreLogs)
	x.IgnoreFiles = copyStringSlice(x.IgnoreFiles)
	x.Theme = copyStringSlice(x.Theme)

	return x
}

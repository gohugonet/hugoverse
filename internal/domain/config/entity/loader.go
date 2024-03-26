package entity

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gohugonet/hugoverse/internal/domain/config"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ConfigLoader struct {
	Path string
}

func (c *ConfigLoader) LoadConfigFromDisk() (map[string]any, error) {
	content, err := os.ReadFile(c.Path)
	if err != nil {
		return nil, err
	}

	configData := bytes.TrimSuffix(content, []byte("\n"))
	format := FormatFromString(path.Base(c.Path))
	m := make(map[string]any)

	if err := UnmarshalTo(configData, format, &m); err != nil {
		return nil, err
	}

	return m, nil
}

// FormatFromString turns formatStr, typically a file extension without any ".",
// into a Format. It returns an empty string for unknown formats.
func FormatFromString(formatStr string) config.Format {
	formatStr = strings.ToLower(formatStr)
	if strings.Contains(formatStr, ".") {
		// Assume a filename
		formatStr = strings.TrimPrefix(
			filepath.Ext(formatStr), ".")
	}
	switch formatStr {
	case "toml":
		return config.TOML
	}

	return ""
}

// UnmarshalTo unmarshal data in format f into v.
func UnmarshalTo(data []byte, f config.Format, v any) error {
	var err error

	switch f {
	case config.TOML:
		err = toml.Unmarshal(data, v)

	default:
		return fmt.Errorf(
			"unmarshal of format %q is not supported", f)
	}

	if err == nil {
		return nil
	}

	return err
}

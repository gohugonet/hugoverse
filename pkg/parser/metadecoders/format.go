package metadecoders

import (
	"path/filepath"
	"strings"
)

type Format string

const (
	// These are the supported metdata  formats in Hugo. Most of these are also
	// supported as /data formats.
	ORG  Format = "org"
	JSON Format = "json"
	TOML Format = "toml"
	YAML Format = "yaml"
	CSV  Format = "csv"
	XML  Format = "xml"
)

// FormatFromString turns formatStr, typically a file extension without any ".",
// into a Format. It returns an empty string for unknown formats.
func FormatFromString(formatStr string) Format {
	formatStr = strings.ToLower(formatStr)
	if strings.Contains(formatStr, ".") {
		// Assume a filename
		formatStr = strings.TrimPrefix(filepath.Ext(formatStr), ".")
	}
	switch formatStr {
	case "toml":
		return TOML
	}

	return ""
}

// FormatFromContentString tries to detect the format (JSON, YAML, TOML or XML)
// in the given string.
// It return an empty string if no format could be detected.
func (d Decoder) FormatFromContentString(data string) Format {
	csvIdx := strings.IndexRune(data, d.Delimiter)
	jsonIdx := strings.Index(data, "{")
	yamlIdx := strings.Index(data, ":")
	xmlIdx := strings.Index(data, "<")
	tomlIdx := strings.Index(data, "=")

	if isLowerIndexThan(csvIdx, jsonIdx, yamlIdx, xmlIdx, tomlIdx) {
		return CSV
	}

	if isLowerIndexThan(jsonIdx, yamlIdx, xmlIdx, tomlIdx) {
		return JSON
	}

	if isLowerIndexThan(yamlIdx, xmlIdx, tomlIdx) {
		return YAML
	}

	if isLowerIndexThan(xmlIdx, tomlIdx) {
		return XML
	}

	if tomlIdx != -1 {
		return TOML
	}

	return ""
}

func isLowerIndexThan(first int, others ...int) bool {
	if first == -1 {
		return false
	}
	for _, other := range others {
		if other != -1 && other < first {
			return false
		}
	}

	return true
}

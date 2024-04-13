package valueobject

type Import struct {
	// Module path
	Path string
	// Set when Path is replaced in project config.
	pathProjectReplaced bool
	// Ignore any config in config.toml (will still follow imports).
	IgnoreConfig bool
	// Do not follow any configured imports.
	IgnoreImports bool
	// Do not mount any folder in this import.
	NoMounts bool
	// Never vendor this import (only allowed in main project).
	NoVendor bool
	// Turn off this module.
	Disable bool
	// File mounts.
	Mounts []Mount
}

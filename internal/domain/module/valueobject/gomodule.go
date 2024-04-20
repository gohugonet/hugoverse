package valueobject

import (
	"golang.org/x/mod/module"
	"strings"
	"time"
)

type goBinaryStatus int

const (
	goBinaryStatusOK goBinaryStatus = iota
	goBinaryStatusNotFound
	goBinaryStatusTooOld
)

const (
	GoModFilename = "go.mod"
	GoSumFilename = "go.sum"
)

type GoModules []*GoModule

type GoModule struct {
	Path     string         // module path
	Version  string         // module version
	Versions []string       // available module versions (with -versions)
	Replace  *GoModule      // replaced by this module
	Time     *time.Time     // time version was created
	Update   *GoModule      // available update, if any (with -u)
	Main     bool           // is this the main module?
	Indirect bool           // is this module only an indirect dependency of main module?
	Dir      string         // directory holding files for this module, if any
	GoMod    string         // path to go.mod file for this module, if any
	Error    *goModuleError // error loading module
}

type goModuleError struct {
	Err string // the error itself
}

func (modules GoModules) GetByPath(p string) *GoModule {
	if modules == nil {
		return nil
	}

	for _, m := range modules {
		if strings.EqualFold(p, m.Path) {
			return m
		}
	}

	return nil
}

func IsProbablyModule(path string) bool {
	return module.CheckPath(path) == nil
}

// In the first iteration of Hugo Modules, we do not support multiple
// major versions running at the same time, so we pick the first (upper most).
// We will investigate namespaces in future versions.
// TODO(bep) add a warning when the above happens.
func pathKey(p string) string {
	prefix, _, _ := module.SplitPathVersion(p)

	return strings.ToLower(prefix)
}

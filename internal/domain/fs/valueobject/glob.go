package valueobject

import (
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/glob"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/spf13/afero"
	"path/filepath"
	"strings"
)

// Glob walks the fs and passes all matches to the handle func.
// The handle func can return true to signal a stop.
func Glob(afs afero.Fs, pattern string, handle func(fi fs.FileMetaInfo) (bool, error)) error {
	pattern = glob.NormalizePathNoLower(pattern)
	if pattern == "" {
		return nil
	}
	root := glob.ResolveRootDir(pattern)
	if !strings.HasPrefix(root, "/") {
		root = "/" + root
	}
	pattern = strings.ToLower(pattern)

	g, err := glob.GetGlob(pattern)
	if err != nil {
		return err
	}

	hasSuperAsterisk := strings.Contains(pattern, "**")
	levels := strings.Count(pattern, "/")

	// Signals that we're done.
	done := errors.New("done")

	wfn := func(p string, info fs.FileMetaInfo, err error) error {
		if err != nil {
			fmt.Println("Glob --- ", p, err)
		}
		p = glob.NormalizePath(p)
		if info.IsDir() {
			if !hasSuperAsterisk {
				// Avoid walking to the bottom if we can avoid it.
				if p != "" && strings.Count(p, "/") >= levels {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if g.Match(p) {
			d, err := handle(info)
			if err != nil {
				return err
			}
			if d {
				return done
			}
		}

		return nil
	}

	w := &Walkway{
		Fs:     afs,
		Root:   root,
		WalkFn: wfn,
		Seen:   make(map[string]bool),

		Log:            loggers.NewDefault(),
		FailOnNotExist: true,
	}

	err = w.Walk()

	if !errors.Is(done, err) {
		return err
	}

	return nil
}

package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/pkg/radixtree"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
)

// NewRootMappingFs creates a new RootMappingFs
// on top of the provided with root mappings with
// some optional metadata about the root.
// Note that From represents a virtual root
// that maps to the actual filename in To.
func newRootMappingFs(afs afero.Fs, rms ...valueobject.RootMapping) *valueobject.RootMappingFs {
	t := radixtree.New()
	realMapToRoot := radixtree.New()
	var virtualRoots []valueobject.RootMapping

	addMapping := func(key string, rm valueobject.RootMapping, to *radixtree.Tree) {
		var mappings []valueobject.RootMapping
		v, found := to.Get(key)
		if found {
			// There may be more than one language pointing to the same root.
			mappings = v.([]valueobject.RootMapping)
		}
		mappings = append(mappings, rm)
		to.Insert(key, mappings)
	}

	for _, rm := range rms {
		fi, err := afs.Stat(rm.To)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil
		}

		meta := rm.Meta.Copy()
		if !fi.IsDir() {
			_, name := filepath.Split(rm.From)
			meta.Name = name
		}

		rm.Fi = valueobject.NewFileMetaInfo(fi, meta)

		key := fs.FilepathSeparator + rm.From // /content
		mappings := valueobject.GetRms(t, key)
		mappings = append(mappings, rm)
		t.Insert(key, mappings)

		rev := rm.To
		if !strings.HasPrefix(rev, fs.FilepathSeparator) {
			rev = fs.FilepathSeparator + rev
		}
		addMapping(rev, rm, realMapToRoot)

		log.Printf("key: %s, mappings: %+v\n", key, mappings)

		virtualRoots = append(virtualRoots, rm)
	}

	// Searching from root
	t.Insert(fs.FilepathSeparator, virtualRoots)

	return &valueobject.RootMappingFs{
		Fs:            afs,
		RootMapToReal: t,
		RealMapToRoot: realMapToRoot,
	}
}

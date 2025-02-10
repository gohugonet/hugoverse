package entity

import (
	"fmt"
	"github.com/mdfriday/hugoverse/internal/domain/resources"
	"github.com/mdfriday/hugoverse/pkg/identity"
	pio "github.com/mdfriday/hugoverse/pkg/io"
	"github.com/mdfriday/hugoverse/pkg/media"
	"io"
	"path"
)

type BundlerClient struct {
	rs *Resources
}

func NewBundlerClient(rs *Resources) *BundlerClient {
	return &BundlerClient{
		rs: rs,
	}
}

// Concat concatenates the list of Resource objects.
func (c *BundlerClient) Concat(targetPath string, r []resources.Resource) (resources.Resource, error) {
	targetPath = path.Clean(targetPath)
	return c.rs.Cache.GetOrCreateResource(targetPath, func() (resources.Resource, error) {
		var resolvedm media.Type

		// The given set of resources must be of the same Media Type.
		// We may improve on that in the future, but then we need to know more.
		for i, rr := range r {
			if i > 0 && rr.MediaType().Type != resolvedm.Type {
				return nil, fmt.Errorf("resources in Concat must be of the same Media Type, got %q and %q", rr.MediaType().Type, resolvedm.Type)
			}
			resolvedm = rr.MediaType()
		}

		idm := identity.NewManager("concat")
		// Add the concatenated resources as dependencies to the composite resource
		// so that we can track changes to the individual resources.
		idm.AddIdentityForEach(identity.ForEeachIdentityProviderFunc(
			func(f func(identity.Identity) bool) bool {
				var terminate bool
				for _, rr := range r {
					identity.WalkIdentitiesShallow(rr, func(depth int, id identity.Identity) bool {
						terminate = f(id)
						return terminate
					})
					if terminate {
						break
					}
				}
				return terminate
			},
		))

		concatr := func() (pio.ReadSeekCloser, error) {
			var rcsources []pio.ReadSeekCloser
			for _, s := range r {
				rcr, ok := s.(resources.ReadSeekCloserResource)
				if !ok {
					return nil, fmt.Errorf("resource %T does not implement resource.ReadSeekerCloserResource", s)
				}
				rc, err := rcr.ReadSeekCloser()
				if err != nil {
					// Close the already opened.
					for _, rcs := range rcsources {
						rcs.Close()
					}
					return nil, err
				}

				rcsources = append(rcsources, rc)
			}

			// Arbitrary JavaScript files require a barrier between them to be safely concatenated together.
			// Without this, the last line of one file can affect the first line of the next file and change how both files are interpreted.
			if resolvedm.MainType == media.Builtin.JavascriptType.MainType && resolvedm.SubType == media.Builtin.JavascriptType.SubType {
				readers := make([]pio.ReadSeekCloser, 2*len(rcsources)-1)
				j := 0
				for i := 0; i < len(rcsources); i++ {
					if i > 0 {
						readers[j] = pio.NewReadSeekerNoOpCloserFromString("\n;\n")
						j++
					}
					readers[j] = rcsources[i]
					j++
				}
				return newMultiReadSeekCloser(readers...), nil
			}

			return newMultiReadSeekCloser(rcsources...), nil
		}

		rsb := newResourceBuilder(targetPath, concatr)
		rsb.withCache(c.rs.Cache).withMediaService(c.rs.MediaService).
			withImageService(c.rs.ImageService).withImageProcessor(c.rs.ImageProc).
			withPublisher(c.rs.Publisher).withURLService(c.rs.URLService)

		return rsb.build()
	})
}

func newMultiReadSeekCloser(sources ...pio.ReadSeekCloser) *multiReadSeekCloser {
	mr := io.MultiReader(toReaders(sources)...)
	return &multiReadSeekCloser{mr, sources}
}

type multiReadSeekCloser struct {
	mr      io.Reader
	sources []pio.ReadSeekCloser
}

func toReaders(sources []pio.ReadSeekCloser) []io.Reader {
	readers := make([]io.Reader, len(sources))
	for i, r := range sources {
		readers[i] = r
	}
	return readers
}

func (r *multiReadSeekCloser) Read(p []byte) (n int, err error) {
	return r.mr.Read(p)
}

func (r *multiReadSeekCloser) Seek(offset int64, whence int) (newOffset int64, err error) {
	for _, s := range r.sources {
		newOffset, err = s.Seek(offset, whence)
		if err != nil {
			return
		}
	}

	r.mr = io.MultiReader(toReaders(r.sources)...)

	return
}

func (r *multiReadSeekCloser) Close() error {
	for _, s := range r.sources {
		s.Close()
	}
	return nil
}

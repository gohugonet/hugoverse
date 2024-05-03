package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
	pio "github.com/gohugonet/hugoverse/pkg/io"
	"image"
	"io"
)

type ImageCache struct {
	rc *ResourceCache

	Fcache *filecache.Cache
	Mcache *dynacache.Partition[string, *resourceAdapter]
}

func NewImageCache(rc *ResourceCache, fileCache *filecache.Cache, memCache *dynacache.Cache) *ImageCache {
	return &ImageCache{
		rc:     rc,
		Fcache: fileCache,
		Mcache: dynacache.GetOrCreatePartition[string, *resourceAdapter](
			memCache,
			"/imgs",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnChange, Weight: 70},
		),
	}
}

func (c *ImageCache) getOrCreate(
	parent *imageResource, conf valueobject.ImageConfig,
	createImage func() (*imageResource, image.Image, error),
) (*resourceAdapter, error) {
	relTarget := parent.relTargetPathFromConfig(conf)
	relTargetPath := relTarget.TargetPath()
	memKey := relTargetPath
	memKey = dynacache.CleanKey(memKey)

	v, err := c.Mcache.GetOrCreate(memKey, func(key string) (*resourceAdapter, error) {
		var img *imageResource

		// These funcs are protected by a named lock.
		// read clones the parent to its new name and copies
		// the content to the destinations.
		read := func(info filecache.ItemInfo, r io.ReadSeeker) error {
			img = parent.clone(nil)
			targetPath := img.getResourcePaths()
			targetPath.File = relTarget.File
			img.SetTargetPath(targetPath)
			img.SetOpenSource(func() (pio.ReadSeekCloser, error) {
				return c.Fcache.Fs.Open(info.Name)
			})
			img.SetSourceFilenameIsHash(true)
			img.SetMediaType(valueobject.MediaType(conf.TargetFormat))

			if err := img.InitConfig(r); err != nil {
				return err
			}

			return nil
		}

		// create creates the image and encodes it to the cache (w).
		create := func(info filecache.ItemInfo, w io.WriteCloser) (err error) {
			defer w.Close()

			var conv image.Image
			img, conv, err = createImage()
			if err != nil {
				return
			}
			targetPath := img.getResourcePaths()
			targetPath.File = relTarget.File
			img.SetTargetPath(targetPath)
			img.SetOpenSource(func() (pio.ReadSeekCloser, error) {
				return c.Fcache.Fs.Open(info.Name)
			})
			return img.EncodeTo(conf, conv, w)
		}

		// Now look in the file cache.

		// The definition of this counter is not that we have processed that amount
		// (e.g. resized etc.), it can be fetched from file cache,
		//  but the count of processed image variations for this site.
		// TODO
		//c.pathSpec.ProcessingStats.Incr(&c.pathSpec.ProcessingStats.ProcessedImages)

		_, err := c.Fcache.ReadOrCreate(relTargetPath, read, create)
		if err != nil {
			return nil, err
		}

		imgAdapter := newResourceAdapter(c.rc, parent.getSpec(), true, img)

		return imgAdapter, nil
	})

	return v, err
}

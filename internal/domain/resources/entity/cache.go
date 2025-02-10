package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/resources"
	"github.com/mdfriday/hugoverse/internal/domain/resources/valueobject"
	"github.com/mdfriday/hugoverse/pkg/cache/dynacache"
	"github.com/mdfriday/hugoverse/pkg/cache/filecache"
	pio "github.com/mdfriday/hugoverse/pkg/io"
	"image"
	"io"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type Cache struct {
	sync.RWMutex

	filecache.Caches

	CacheImage                  *dynacache.Partition[string, *ResourceImage]
	CacheResource               *dynacache.Partition[string, resources.Resource]
	CacheResources              *dynacache.Partition[string, []resources.Resource]
	CacheResourceTransformation *dynacache.Partition[string, *Resource]
}

func (c *Cache) GetOrCreateResource(key string, f func() (resources.Resource, error)) (resources.Resource, error) {
	return c.CacheResource.GetOrCreate(key, func(key string) (resources.Resource, error) {
		return f()
	})
}

func (c *Cache) GetOrCreateResources(key string, f func() ([]resources.Resource, error)) ([]resources.Resource, error) {
	return c.CacheResources.GetOrCreate(key, func(key string) ([]resources.Resource, error) {
		return f()
	})
}

func (c *Cache) GetOrCreateImageResource(parent *ResourceImage, conf valueobject.ImageConfig,
	createImage func() (*ResourceImage, image.Image, error)) (*ResourceImage, error) {

	relTarget := parent.relTargetPathFromConfig(conf)
	relTargetPath := relTarget.TargetPath()
	memKey := relTargetPath
	memKey = dynacache.CleanKey(memKey)

	v, err := c.CacheImage.GetOrCreate(memKey, func(key string) (*ResourceImage, error) {
		var img *ResourceImage

		// These funcs are protected by a named lock.
		// read clones the parent to its new name and copies
		// the content to the destinations.
		read := func(info filecache.ItemInfo, r io.ReadSeeker) error {
			img = parent.clone(nil)
			targetPath := img.paths
			targetPath.File = relTarget.File
			img.paths = targetPath
			img.Resource.openReadSeekCloser = func() (pio.ReadSeekCloser, error) {
				return c.Caches.ImageCache().Fs.Open(info.Name)
			}
			img.Resource.mediaType = valueobject.MediaType(conf.TargetFormat)

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
			targetPath := img.paths
			targetPath.File = relTarget.File
			img.paths = targetPath
			img.Resource.openReadSeekCloser = func() (pio.ReadSeekCloser, error) {
				return c.Caches.ImageCache().Fs.Open(info.Name)
			}
			return img.EncodeTo(conf, conv, w)
		}

		// Now look in the file cache.

		// The definition of this counter is not that we have processed that amount
		// (e.g. resized etc.), it can be fetched from file cache,
		//  but the count of processed image variations for this site.
		// TODO
		//c.pathSpec.ProcessingStats.Incr(&c.pathSpec.ProcessingStats.ProcessedImages)

		_, err := c.Caches.ImageCache().ReadOrCreate(relTargetPath, read, create)
		if err != nil {
			return nil, err
		}

		return img, nil
	})

	return v, err
}

func (c *Cache) CleanKey(key string) string {
	return strings.TrimPrefix(path.Clean(strings.ToLower(filepath.ToSlash(key))), "/")
}

// WriteMeta writes the metadata to file and returns a writer for the content part.
func (c *Cache) WriteMeta(key string, metaRaw []byte) (filecache.ItemInfo, io.WriteCloser, error) {
	filenameMeta, filenameContent := c.getFilenames(key)

	_, fm, err := c.Caches.AssetsCache().WriteCloser(filenameMeta)
	if err != nil {
		return filecache.ItemInfo{}, nil, err
	}
	defer fm.Close()

	if _, err := fm.Write(metaRaw); err != nil {
		return filecache.ItemInfo{}, nil, err
	}

	fi, fc, err := c.Caches.AssetsCache().WriteCloser(filenameContent)

	return fi, fc, err
}

func (c *Cache) getFilenames(key string) (string, string) {
	filenameMeta := key + ".json"
	filenameContent := key + ".Content"

	return filenameMeta, filenameContent
}

func (c *Cache) GetFile(key string) (filecache.ItemInfo, io.ReadCloser, []byte, bool) {
	c.RLock()
	defer c.RUnlock()

	filenameMeta, filenameContent := c.getFilenames(key)

	_, jsonContent, _ := c.Caches.AssetsCache().GetBytes(filenameMeta)

	fi, rc, _ := c.Caches.AssetsCache().Get(filenameContent)

	return fi, rc, jsonContent, rc != nil
}

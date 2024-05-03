package factory

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/entity"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/cache/dynacache"
	"github.com/gohugonet/hugoverse/pkg/cache/filecache"
	"github.com/gohugonet/hugoverse/pkg/hexec"
	"github.com/gohugonet/hugoverse/pkg/identity"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/resource/jsconfig"
	"github.com/spf13/afero"
	"net/http"
	"time"
)

func NewResources(ws resources.Workspace) (resources.Resources, error) {
	fileCaches, err := newCaches(ws)
	if err != nil {
		return nil, err
	}
	memoryCache := newMemoryCache()
	resourceCache := newResourceCache(fileCaches.AssetsCache(), memoryCache)

	execHelper := newExecHelper(ws)
	ip, err := newImageProcessor(ws)
	if err != nil {
		return nil, err
	}

	ic := entity.NewImageCache(
		resourceCache,
		fileCaches.ImageCache(),
		memoryCache,
	)

	common := &entity.Common{
		Incr:       &identity.IncrementByOne{},
		FileCaches: fileCaches,
		PostBuildAssets: &entity.PostBuildAssets{
			PostProcessResources: make(map[string]resources.PostPublishedResource),
			JSConfigBuilder:      jsconfig.NewBuilder(),
		},
	}

	rs := &entity.Resources{
		Creator: &entity.Creator{
			MediaService: ws,
			UrlService:   ws,
			GlobService:  ws,

			AssetsFs:  ws.AssetsFs(),
			PublishFs: ws.PublishFs(),

			HttpClient: &http.Client{
				Timeout: time.Minute,
			},
			CacheGetResource: fileCaches.GetResourceCache(),
			ResourceCache:    resourceCache,

			Imaging:    ip,
			ImageCache: ic,
		},
		ImageCache: ic,
		ExecHelper: execHelper,
		Common:     common,
	}

	return rs, nil
}

func newCaches(ws resources.Workspace) (filecache.Caches, error) {
	fs := ws.SourceFs()

	m := make(filecache.Caches)
	ws.CachesIterator(func(cacheKey string, isResourceDir bool, dir string, age time.Duration) error {
		var cfs afero.Fs

		if isResourceDir {
			cfs = ws.ResourcesCacheFs()
		} else {
			cfs = fs
		}

		if cfs == nil {
			panic("nil fs")
		}

		baseDir := dir

		bfs := ws.NewBasePathFs(cfs, baseDir)

		var pruneAllRootDir string
		if cacheKey == "modules" {
			pruneAllRootDir = "pkg"
		}

		m[cacheKey] = filecache.NewCache(bfs, age, pruneAllRootDir)
		return nil
	})

	return m, nil
}

func newMemoryCache() *dynacache.Cache {
	return dynacache.New(dynacache.Options{Running: true, Log: loggers.NewDefault()})
}

func newExecHelper(ws resources.Workspace) *hexec.Exec {
	return hexec.NewWithAuth(ws.ExecAuth())
}

func newImageProcessor(ws resources.Workspace) (*valueobject.ImageProcessor, error) {
	exifDecoder, err := ws.ExifDecoder()
	if err != nil {
		return nil, err
	}
	return &valueobject.ImageProcessor{
		ExifDecoder: exifDecoder,
	}, nil
}

func newResourceCache(assetsCache *filecache.Cache, memCache *dynacache.Cache) *entity.ResourceCache {
	return &entity.ResourceCache{
		FileCache: assetsCache,
		CacheResource: dynacache.GetOrCreatePartition[string, resources.Resource](
			memCache,
			"/res1",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnChange, Weight: 40},
		),
		CacheResources: dynacache.GetOrCreatePartition[string, []resources.Resource](
			memCache,
			"/ress",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnRebuild, Weight: 40},
		),
		CacheResourceTransformation: dynacache.GetOrCreatePartition[string, *entity.ResourceAdapterInner](
			memCache,
			"/res1/tra",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnChange, Weight: 40},
		),
	}
}

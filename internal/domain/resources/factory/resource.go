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
	"time"
)

func NewResources(ws resources.Workspace) (*entity.Resources, error) {
	c, err := newCache(ws)
	if err != nil {
		return nil, err
	}

	execHelper := newExecHelper(ws)
	log := loggers.NewDefault()
	ds, err := newDartSass(execHelper, ws)
	if err != nil {
		log.Errorln(err)
	}

	ip, err := newImageProcessor(ws)
	if err != nil {
		return nil, err
	}

	mc, err := NewMinifier(ws.MediaTypes(), ws.AllOutputFormats(), ws)
	if err != nil {
		return nil, err
	}

	common := &entity.Common{
		Incr: &identity.IncrementByOne{},
		PostBuildAssets: &entity.PostBuildAssets{
			PostProcessResources: make(map[string]resources.PostPublishedResource),
			JSConfigBuilder:      jsconfig.NewBuilder(),
		},
	}

	rs := &entity.Resources{
		Cache:        c,
		FsService:    ws,
		MediaService: ws,
		UrlService:   ws,

		ImageService: ws,
		ImageProc:    ip,

		ExecHelper: execHelper,
		Common:     common,

		MinifierClient:  mc,
		TemplateClient:  nil,
		IntegrityClient: &entity.IntegrityClient{},
		SassClient:      ds,
	}

	return rs, nil
}

func newCache(ws resources.Workspace) (*entity.Cache, error) {
	fileCaches, err := newCaches(ws)
	if err != nil {
		return nil, err
	}
	memoryCache := newMemoryCache()

	return &entity.Cache{
		Caches: fileCaches,
		CacheImage: dynacache.GetOrCreatePartition[string, *entity.ResourceImage](
			memoryCache,
			"/imgs",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnChange, Weight: 70},
		),
		CacheResource: dynacache.GetOrCreatePartition[string, resources.Resource](
			memoryCache,
			"/res1",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnChange, Weight: 40},
		),
		CacheResources: dynacache.GetOrCreatePartition[string, []resources.Resource](
			memoryCache,
			"/ress",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnRebuild, Weight: 40},
		),
		CacheResourceTransformation: dynacache.GetOrCreatePartition[string, *entity.Resource](
			memoryCache,
			"/res1/tra",
			dynacache.OptionsPartition{ClearWhen: dynacache.ClearOnChange, Weight: 40},
		),
	}, nil
}

func newCaches(ws resources.Workspace) (filecache.Caches, error) {
	fs := ws.SourceFs()

	m := make(filecache.Caches)
	ws.CachesIterator(func(cacheKey string, isResourceDir bool, dir string, age time.Duration) {
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

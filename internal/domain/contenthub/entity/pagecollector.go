package entity

import (
	"context"
	"fmt"
	"github.com/bep/logg"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/pkg/env"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"github.com/gohugonet/hugoverse/pkg/rungroup"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type pagesCollector struct {
	m  *PageMap
	fs contenthub.FsService

	infoLogger logg.LevelLogger

	ctx context.Context
	g   rungroup.Group[fs.FileMetaInfo]
}

// Collect pages.
func (c *pagesCollector) Collect() (collectErr error) {
	var (
		numWorkers             = env.GetNumWorkerMultiplier()
		numFilesProcessedTotal atomic.Uint64
		numFilesProcessedLast  uint64
		fileBatchTimer         = time.Now()
		fileBatchTimerMu       sync.Mutex
	)

	l := c.infoLogger.WithField("subStep", "collect")
	logFilesProcessed := func(force bool) {
		fileBatchTimerMu.Lock()
		if force || time.Since(fileBatchTimer) > 3*time.Second {
			numFilesProcessedBatch := numFilesProcessedTotal.Load() - numFilesProcessedLast
			numFilesProcessedLast = numFilesProcessedTotal.Load()
			loggers.TimeTrackf(l, fileBatchTimer,
				logg.Fields{
					logg.Field{Name: "files", Value: numFilesProcessedBatch},
					logg.Field{Name: "files_total", Value: numFilesProcessedTotal.Load()},
				},
				"",
			)
			fileBatchTimer = time.Now()
		}
		fileBatchTimerMu.Unlock()
	}
	defer func() {
		logFilesProcessed(true)
	}()

	c.g = rungroup.Run[fs.FileMetaInfo](c.ctx, rungroup.Config[fs.FileMetaInfo]{
		NumWorkers: numWorkers,
		Handle: func(ctx context.Context, fi fs.FileMetaInfo) error {
			if err := c.m.AddFi(fi); err != nil {
				return fs.AddFileInfoToError(err, fi, c.fs)
			}
			numFilesProcessedTotal.Add(1)
			if numFilesProcessedTotal.Load()%1000 == 0 {
				logFilesProcessed(false)
			}
			return nil
		},
	})

	collectErr = c.collectDir()

	werr := c.g.Wait()
	if collectErr == nil {
		collectErr = werr
	}

	return
}

func (c *pagesCollector) collectDir() error {
	dirPath := ""

	root, err := c.fs.ContentFs().Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := c.collectDirDir(dirPath, root.(fs.FileMetaInfo)); err != nil {
		return err
	}

	return nil
}

func (c *pagesCollector) collectDirDir(path string, root fs.FileMetaInfo) error {
	preHook := func(dir fs.FileMetaInfo, path string, readdir []fs.FileMetaInfo) ([]fs.FileMetaInfo, error) {
		if len(readdir) == 0 {
			return nil, nil
		}

		// Pick the first regular file.
		var first fs.FileMetaInfo
		for _, fi := range readdir {
			if fi.IsDir() {
				continue
			}
			first = fi
			break
		}

		if first == nil {
			// Only dirs, keep walking.
			return readdir, nil
		}

		// Any bundle file will always be first.
		firstPi := first.Path()
		if firstPi == nil {
			panic(fmt.Sprintf("collectDirDir: no path info for %q", first.FileName()))
		}

		if firstPi.IsLeafBundle() {
			if err := c.handleBundleLeaf(dir, first, path, readdir); err != nil {
				return nil, err
			}
			return nil, filepath.SkipDir
		}

		for _, fi := range readdir {
			if fi.IsDir() {
				continue
			}

			if err := c.g.Enqueue(fi); err != nil {
				return nil, err
			}
		}

		// Keep walking.
		return readdir, nil
	}

	wfn := func(path string, fi fs.FileMetaInfo) error {
		return nil
	}

	if err := c.fs.WalkContent(path, fs.WalkCallback{
		HookPre:  preHook,
		WalkFn:   wfn,
		HookPost: nil,
	}, fs.WalkwayConfig{}); err != nil {
		return err
	}

	return nil
}

func (c *pagesCollector) handleBundleLeaf(dir, bundle fs.FileMetaInfo, inPath string, readdir []fs.FileMetaInfo) error {
	walk := func(path string, info fs.FileMetaInfo) error {
		if info.IsDir() {
			return nil
		}

		pi := info.Path()

		if info != bundle {
			// Everything inside a leaf bundle is a Resource,
			// even the content pages.
			// Note that we do allow index.md as page resources, but not in the bundle root.
			if !pi.IsLeafBundle() || pi.Dir() != bundle.Path().Dir() {
				paths.ModifyPathBundleTypeResource(pi)
			}
		}

		return c.g.Enqueue(info)
	}

	if err := c.fs.WalkContent(inPath, fs.WalkCallback{
		HookPre:  nil,
		WalkFn:   walk,
		HookPost: nil,
	}, fs.WalkwayConfig{
		Info:       dir,
		DirEntries: readdir,
	}); err != nil {
		return err
	}

	return nil
}

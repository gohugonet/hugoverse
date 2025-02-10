package entity

import (
	"context"
	"github.com/bep/logg"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub"
	"github.com/mdfriday/hugoverse/internal/domain/contenthub/valueobject"
	"github.com/mdfriday/hugoverse/internal/domain/fs"
	"github.com/mdfriday/hugoverse/pkg/env"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/mdfriday/hugoverse/pkg/rungroup"
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
	g   rungroup.Group[*valueobject.File]
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

	c.g = rungroup.Run[*valueobject.File](c.ctx, rungroup.Config[*valueobject.File]{
		NumWorkers: numWorkers,
		Handle: func(ctx context.Context, fi *valueobject.File) error {
			if err := c.m.AddFi(fi); err != nil {
				return valueobject.AddFileInfoToError(err, fi.FileMetaInfo, c.fs.ContentFs())
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

		fsm := valueobject.NewFsManager(readdir)

		leaf := fsm.GetLeaf()
		if leaf != nil {
			c.infoLogger.Logf("handleBundleLeaf: %s", leaf.Name())

			if err := c.handleBundleLeaf(dir, leaf, path, readdir); err != nil {
				return nil, err
			}
			return nil, filepath.SkipDir
		}

		for _, fi := range readdir {
			if fi.IsDir() {
				continue
			}

			file, err := valueobject.NewFileInfo(fi)
			if err != nil {
				return nil, err
			}
			if err := c.g.Enqueue(file); err != nil {
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
	}, fs.WalkwayConfig{Info: root}); err != nil {
		return err
	}

	return nil
}

func (c *pagesCollector) handleBundleLeaf(dir fs.FileMetaInfo, bundle *valueobject.File, inPath string, readdir []fs.FileMetaInfo) error {
	bundlePath := bundle.Paths()

	walk := func(path string, info fs.FileMetaInfo) error {
		if info.IsDir() {
			return nil
		}

		f, err := valueobject.NewFileInfo(info)
		if err != nil {
			return err
		}

		if info != bundle.FileMetaInfo {
			// Everything inside a leaf bundle is a Resource,
			// even the content pages.
			// Note that we do allow index.md as page resources, but not in the bundle root.
			if !f.IsLeafBundle() || f.Paths().Dir() != bundlePath.Dir() {
				f.ShiftToResource()
			}
		}

		return c.g.Enqueue(f)
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

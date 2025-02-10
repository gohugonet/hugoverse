package static

import (
	"github.com/mdfriday/hugoverse/pkg/helpers"
	"github.com/mdfriday/hugoverse/pkg/herrors"
	"github.com/mdfriday/hugoverse/pkg/loggers"
	"github.com/spf13/afero"
	"github.com/spf13/fsync"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Static struct {
	source []afero.Fs
	target afero.Fs
	logger loggers.Logger
}

func Copy(fs []afero.Fs, publish afero.Fs) error {
	s := &Static{
		source: fs,
		target: publish,
		logger: loggers.NewDefault(),
	}

	m, err := s.doWithPublishDirs(s.copyStaticTo)
	if err == nil || herrors.IsNotExist(err) {
		s.logger.Infoln("Copied", m, "static files", "from", s.source[0].Name(), "to", s.target.Name())
	}

	if err != nil {
		s.logger.Errorln(err)
		return err
	}

	return nil
}

func (s *Static) doWithPublishDirs(f func(sourceFs afero.Fs) (uint64, error)) (uint64, error) {
	var count uint64 = 0
	staticFilesystems := s.source

	if len(staticFilesystems) == 0 {
		s.logger.Infoln("No static directories found to sync")
		return count, nil
	}

	for _, fs := range staticFilesystems {
		cnt, err := f(fs)
		if err != nil {
			return count, err
		}
		count += cnt
	}

	return count, nil
}

func (s *Static) copyStaticTo(sourceFs afero.Fs) (uint64, error) {
	infol := s.logger.InfoCommand("static")
	publishDir := helpers.FilePathSeparator

	fs := &countingStatFs{Fs: sourceFs}

	syncer := fsync.NewSyncer()
	syncer.NoTimes = false
	syncer.NoChmod = false
	syncer.ChmodFilter = chmodFilter

	syncer.DestFs = s.target
	// Now that we are using a unionFs for the static directories
	// We can effectively clean the publishDir on initial sync
	syncer.Delete = false

	syncer.SrcFs = fs

	if syncer.Delete {
		infol.Logf("removing all files from destination that don't exist in static dirs")

		syncer.DeleteFilter = func(f fsync.FileInfo) bool {
			return f.IsDir() && strings.HasPrefix(f.Name(), ".")
		}
	}
	start := time.Now()

	// because we are using a baseFs (to get the union right).
	// set sync src to root
	err := syncer.Sync(publishDir, helpers.FilePathSeparator)
	if err != nil {
		return 0, err
	}
	loggers.TimeTrackf(infol, start, nil, "syncing static files to %s", publishDir)

	// Sync runs Stat 2 times for every source file.
	numFiles := fs.statCounter / 2

	return numFiles, err
}

func chmodFilter(dst, src os.FileInfo) bool {
	// Hugo publishes data from multiple sources, potentially
	// with overlapping directory structures. We cannot sync permissions
	// for directories as that would mean that we might end up with write-protected
	// directories inside /public.
	// One example of this would be syncing from the Go Module cache,
	// which have 0555 directories.
	return src.IsDir()
}

type countingStatFs struct {
	afero.Fs
	statCounter uint64
}

func (fs *countingStatFs) Stat(name string) (os.FileInfo, error) {
	f, err := fs.Fs.Stat(name)
	if err == nil {
		if !f.IsDir() {
			atomic.AddUint64(&fs.statCounter, 1)
		}
	}
	return f, err
}

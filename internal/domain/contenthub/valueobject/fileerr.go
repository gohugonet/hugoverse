package valueobject

import (
	"errors"
	"github.com/mdfriday/hugoverse/internal/domain/fs"
	"github.com/mdfriday/hugoverse/pkg/herrors"
	"github.com/spf13/afero"
)

// AddFileInfoToError adds file info to the given error.
func AddFileInfoToError(err error, fi fs.FileMetaInfo, fs afero.Fs) error {
	if err == nil {
		return nil
	}

	filename := fi.FileName()

	// Check if it's already added.
	for _, ferr := range herrors.UnwrapFileErrors(err) {
		pos := ferr.Position()
		errfilename := pos.Filename
		if errfilename == "" {
			pos.Filename = filename
			ferr.UpdatePosition(pos)
		}

		if errfilename == "" || errfilename == filename {
			if filename != "" && ferr.ErrorContext() == nil {
				f, ioerr := fs.Open(filename)
				if ioerr != nil {
					return err
				}
				defer f.Close()
				ferr.UpdateContent(f, nil)
			}
			return err
		}
	}

	lineMatcher := herrors.NopLineMatcher

	var textSegmentErr *herrors.TextSegmentError
	if errors.As(err, &textSegmentErr) {
		lineMatcher = herrors.ContainsMatcher(textSegmentErr.Segment)
	}

	return herrors.NewFileErrorFromFile(err, filename, fs, lineMatcher)
}

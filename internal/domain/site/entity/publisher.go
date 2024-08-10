package entity

import (
	"bytes"
	"errors"
	"github.com/gohugonet/hugoverse/internal/domain/site"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/spf13/afero"
	"io"
	"os"
	"path/filepath"
)

type Publisher struct {
	Fs afero.Fs
}

func (p *Publisher) PublishSource(src *bytes.Buffer, filename ...string) error {
	fw, err := helpers.OpenFilesForWriting(p.Fs, filename...)
	if err != nil {
		return err
	}
	defer fw.Close()

	_, err = io.Copy(fw, src)
	if err != nil {
		return err
	}

	return nil
}

func (p *Publisher) PublishFiles(fr io.ReadSeekCloser, filenames ...string) error {
	var fw io.WriteCloser
	fw, err := helpers.OpenFilesForWriting(p.Fs, filenames...)
	if err != nil {
		return err
	}
	defer fw.Close()

	if _, err = io.Copy(fw, fr); err != nil {
		return err
	}

	return nil
}

// Publish applies any relevant transformations and writes the file
// to its destination, e.g. /public.
func (p *Publisher) Publish(d site.Descriptor) error {
	if d.TargetPath == "" {
		return errors.New("publish: must provide a TargetPath")
	}
	src := d.Src

	f, err := OpenFileForWriting(p.Fs, d.TargetPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var w io.Writer = f

	_, err = io.Copy(w, src)

	return err
}

// OpenFileForWriting opens or creates the given file. If the target directory
// does not exist, it gets created.
func OpenFileForWriting(fs afero.Fs, filename string) (afero.File, error) {
	filename = filepath.Clean(filename)
	// Create will truncate if file already exists.
	// os.Create will create any new files with mode 0666 (before umask).
	f, err := fs.Create(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err = fs.MkdirAll(filepath.Dir(filename), 0777); err != nil { //  before umask
			return nil, err
		}
		f, err = fs.Create(filename)
	}

	return f, err
}

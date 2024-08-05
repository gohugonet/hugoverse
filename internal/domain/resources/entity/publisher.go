package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/helpers"
	"github.com/spf13/afero"
	"io"
)

type Publisher struct {
	PubFs afero.Fs
}

func (p *Publisher) PublishContentToTarget(content, target string) error {
	f, err := p.OpenPublishFileForWriting(target)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(content))
	return err
}

func (p *Publisher) OpenPublishFileForWriting(relTargetPath string) (io.WriteCloser, error) {
	filenames := valueobject.NewResourcePaths(relTargetPath).TargetFilenames()
	return helpers.OpenFilesForWriting(p.PubFs, filenames...)
}

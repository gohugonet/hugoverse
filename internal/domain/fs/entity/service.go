package entity

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
)

type Service struct {
}

func (s *Service) NewFileMetaInfo(filename string) fs.FileMetaInfo {
	return valueobject.NewFileInfoWithName(filename)
}

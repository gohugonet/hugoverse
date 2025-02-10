package entity

import (
	"github.com/mdfriday/hugoverse/internal/domain/fs"
	"github.com/mdfriday/hugoverse/internal/domain/fs/valueobject"
)

type Service struct {
}

func (s *Service) NewFileMetaInfo(filename string) fs.FileMetaInfo {
	return valueobject.NewFileInfoWithName(filename)
}

func (s *Service) NewFileMetaInfoWithContent(content string) fs.FileMetaInfo {
	return valueobject.NewFileInfoWithContent(content)
}

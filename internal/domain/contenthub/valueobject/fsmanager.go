package valueobject

import (
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	"sort"
)

type FileManager struct {
	files []*File
}

func NewFsManager(fis []fs.FileMetaInfo) *FileManager {
	var fsFiles []*File
	for _, f := range fis {
		fsFiles = append(fsFiles, NewFileInfo(f))
	}

	fsm := &FileManager{
		files: fsFiles,
	}

	fsm.sort()

	return fsm
}

func (f *FileManager) sort() {
	sort.Slice(f.files, func(i, j int) bool {
		fi, fj := f.files[i], f.files[j]

		fimi, fimj := fi.FileMetaInfo, fj.FileMetaInfo
		if fimi.IsDir() != fimj.IsDir() {
			return fimi.IsDir()
		}

		pii, pij := f.files[i].Path(), f.files[j].Path()
		if pii != nil {
			// Pull bundles to the top.
			if fi.IsBundle() != fj.IsBundle() {
				return fi.IsBundle()
			}

			exti, extj := fi.Ext(), fj.Ext()
			if exti != extj {
				// This pulls .md above .html.
				return exti > extj
			}

			basei, basej := pii.Base(), pij.Base()
			if basei != basej {
				return basei < basej
			}
		}

		return fimi.Name() < fimj.Name()
	})
}

func (f *FileManager) GetLeaf() *File {
	for _, f := range f.files {
		if f.IsLeafBundle() {
			return f
		}
	}
	return nil
}

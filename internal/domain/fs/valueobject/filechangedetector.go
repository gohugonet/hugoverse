package valueobject

import (
	"regexp"
	"sync"
)

type FileChangeDetector struct {
	sync.Mutex
	current map[string]string
	prev    map[string]string

	IrrelevantRe *regexp.Regexp
}

func NewFileChangeDetector() *FileChangeDetector {
	return &FileChangeDetector{
		current:      map[string]string{},
		prev:         map[string]string{},
		IrrelevantRe: regexp.MustCompile(`\.map$`),
	}
}

func (f *FileChangeDetector) OnFileClose(name, md5sum string) {
	f.Lock()
	defer f.Unlock()
	f.current[name] = md5sum
}

func (f *FileChangeDetector) PrepareNew() {
	if f == nil {
		return
	}

	f.Lock()
	defer f.Unlock()

	if f.current == nil {
		f.current = make(map[string]string)
		f.prev = make(map[string]string)
		return
	}

	f.prev = make(map[string]string)
	for k, v := range f.current {
		f.prev[k] = v
	}
	f.current = make(map[string]string)
}

func (f *FileChangeDetector) changed() []string {
	if f == nil {
		return nil
	}
	f.Lock()
	defer f.Unlock()
	var c []string
	for k, v := range f.current {
		vv, found := f.prev[k]
		if !found || v != vv {
			c = append(c, k)
		}
	}

	return f.filterIrrelevant(c)
}

func (f *FileChangeDetector) filterIrrelevant(in []string) []string {
	var filtered []string
	for _, v := range in {
		if !f.IrrelevantRe.MatchString(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

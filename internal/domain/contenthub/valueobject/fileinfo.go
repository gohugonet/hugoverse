package valueobject

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/contenthub"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	fsVO "github.com/gohugonet/hugoverse/internal/domain/fs/valueobject"
	"github.com/gohugonet/hugoverse/pkg/paths"
	"path/filepath"
	"strings"
	"sync"
)

type File struct {
	contenthub.File
}

func NewFileInfo(fi fs.FileMetaInfo) (*File, error) {
	baseFi, err := newFileInfo(fi)
	if err != nil {
		return nil, err
	}

	f := &File{
		File: baseFi,
	}

	return f, nil
}

func newFileInfo(fi fs.FileMetaInfo) (*FileInfo, error) {
	m := fi.Meta()

	filename := m.Filename
	relPath := m.Path

	if relPath == "" {
		return nil, fmt.Errorf("no Path provided by %v (%T)", m, m.Fs)
	}

	if filename == "" {
		return nil, fmt.Errorf("no Filename provided by %v (%T)", m, m.Fs)
	}

	relDir := filepath.Dir(relPath)
	if relDir == "." {
		relDir = ""
	}
	if !strings.HasSuffix(relDir, paths.FilePathSeparator) {
		relDir = relDir + paths.FilePathSeparator
	}

	dir, name := filepath.Split(relPath)
	if !strings.HasSuffix(dir, paths.FilePathSeparator) {
		dir = dir + paths.FilePathSeparator
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	baseName := paths.Filename(name)

	f := &FileInfo{
		filename:   filename,
		fi:         fi,
		ext:        ext,
		dir:        dir,
		relDir:     relDir,  // Dir()
		relPath:    relPath, // Path()
		name:       name,
		baseName:   baseName, // BaseFileName()
		classifier: m.Classifier,
	}

	return f, nil
}

// FileInfo describes a source file.
type FileInfo struct {

	// Absolute filename to the file on disk.
	filename string

	fi fsVO.FileMetaInfo

	// Derived from filename
	ext string // Extension without any "."

	name string

	dir                 string
	relDir              string
	relPath             string
	baseName            string
	translationBaseName string
	contentBaseName     string
	section             string
	classifier          fsVO.ContentClass

	uniqueID string

	lazyInit sync.Once
}

// Path gets the relative path including file name and extension.  The directory
// is relative to the content root.
func (fi *FileInfo) Path() string { return fi.relPath }

// Section returns a file's section.
func (fi *FileInfo) Section() string {
	fi.init()
	return fi.section
}

// We create a lot of these FileInfo objects, but there are parts of it used only
// in some cases that is slightly expensive to construct.
func (fi *FileInfo) init() {
	fi.lazyInit.Do(func() {
		relDir := strings.Trim(fi.relDir, paths.FilePathSeparator)
		parts := strings.Split(relDir, paths.FilePathSeparator)
		var section string
		if len(parts) > 1 {
			section = parts[0]
		}
		fi.section = section
		fi.contentBaseName = fi.translationBaseName
		fi.uniqueID = MD5String(filepath.ToSlash(fi.relPath))
	})
}

// MD5String takes a string and returns its MD5 hash.
func MD5String(f string) string {
	h := md5.New()
	h.Write([]byte(f))
	return hex.EncodeToString(h.Sum([]byte{}))
}

func (fi *FileInfo) IsZero() bool {
	return fi == nil
}

// Dir gets the name of the directory that contains this file.  The directory is
// relative to the content root.
func (fi *FileInfo) Dir() string { return fi.relDir }

// Extension is an alias to Ext().
func (fi *FileInfo) Extension() string {
	return fi.Ext()
}

// Ext returns a file's extension without the leading period (ie. "md").
func (fi *FileInfo) Ext() string { return fi.ext }

// Filename returns a file's absolute path and filename on disk.
func (fi *FileInfo) Filename() string { return fi.filename }

// LogicalName returns a file's name and extension (ie. "page.sv.md").
func (fi *FileInfo) LogicalName() string { return fi.name }

// BaseFileName returns a file's name without extension (ie. "page.sv").
func (fi *FileInfo) BaseFileName() string { return fi.baseName }

// TranslationBaseName returns a file's translation base name without the
// language segment (ie. "page").
func (fi *FileInfo) TranslationBaseName() string { return fi.translationBaseName }

// ContentBaseName is a either TranslationBaseName or name of containing folder
// if file is a leaf bundle.
func (fi *FileInfo) ContentBaseName() string {
	fi.init()
	return fi.contentBaseName
}

// UniqueID returns a file's unique, MD5 hash identifier.
func (fi *FileInfo) UniqueID() string {
	fi.init()
	return fi.uniqueID
}

// FileInfo returns a file's underlying os.FileInfo.
func (fi *FileInfo) FileInfo() fsVO.FileMetaInfo { return fi.fi }

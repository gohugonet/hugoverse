package search

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"log"
	"os"
	"path/filepath"
)

var (
	// Search tracks all search indices to use throughout system
	Search map[string]bleve.Index

	// ErrNoIndex is for failed checks for an index in Search map
	ErrNoIndex = errors.New("no search index found for type provided")

	ContentTypes map[string]content.Creator
)

// Setup initializes Search Index for search to be functional
// This was moved out of db.Init and put to main(), because addon checker was initializing db together with
// search indexing initialisation in time when there were no item.Types defined so search index was always
// empty when using addons. We still have no guarentee whatsoever that item.Types is defined
// Should be called from a goroutine after SetContent is successful (SortContent requirement)
func Setup(cts map[string]content.Creator, searchDir string) {
	Search = make(map[string]bleve.Index)
	ContentTypes = cts

	for t := range ContentTypes {
		err := MapIndex(t, searchDir)
		if err != nil {
			log.Fatalln(err)
			return
		}
	}
}

// MapIndex creates the mapping for a type and tracks the index to be used within
// the system for adding/deleting/checking data
func MapIndex(typeName string, searchDir string) error {
	// type assert for Searchable, get configuration (which can be overridden)
	// by Ponzu user if defines own SearchMapping()
	it, ok := ContentTypes[typeName]
	if !ok {
		return fmt.Errorf("[search] MapIndex Error: Failed to MapIndex for %s, type doesn't exist", typeName)
	}
	s, ok := it().(content.Searchable)
	if !ok {
		return fmt.Errorf("[search] MapIndex Error: Item type %s doesn't implement search.Searchable", typeName)
	}

	// skip setting or using index for types that shouldn't be indexed
	if !s.IndexContent() {
		fmt.Printf("[search] Index not created for %s\n", typeName)
		return nil
	}

	mapping, err := s.SearchMapping()
	if err != nil {
		fmt.Println(err)
		return err
	}

	idxName := typeName + ".index"
	var idx bleve.Index

	searchPath := searchDir

	err = os.MkdirAll(searchPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	idxPath := filepath.Join(searchPath, idxName)
	if _, err = os.Stat(idxPath); os.IsNotExist(err) {
		idx, err = bleve.New(idxPath, mapping)
		if err != nil {
			return err
		}
		idx.SetName(idxName)
	} else {
		idx, err = bleve.Open(idxPath)
		if err != nil {
			return err
		}
	}

	// add the type name to the index and track the index
	Search[typeName] = idx
	fmt.Printf("[search] Index created for %s\n", typeName)

	return nil
}

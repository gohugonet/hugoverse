package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	// SearchIndices tracks all search indices to use throughout system
	SearchIndices map[string]bleve.Index

	// ErrNoIndex is for failed checks for an index in Search map
	ErrNoIndex = errors.New("no search index found for type provided")

	ContentTypes map[string]content.Creator
)

type Search struct{}

// TypeQuery conducts a search and returns a set of Ponzu "targets", Type:ID pairs,
// and an error. If there is no search index for the typeName (Type) provided,
// db.ErrNoIndex will be returned as the error
func (s *Search) TypeQuery(typeName, query string, count, offset int) ([]content.Identifier, error) {
	idx, ok := SearchIndices[typeName]
	if !ok {
		fmt.Println("Index for type ", typeName, " not found")
		return nil, ErrNoIndex
	}

	q := bleve.NewQueryStringQuery(query)
	req := bleve.NewSearchRequestOptions(q, count, offset, false)
	res, err := idx.Search(req)
	if err != nil {
		return nil, err
	}

	var results []content.Identifier
	for _, hit := range res.Hits {
		results = append(results, valueobject.CreateIndex(hit.ID))
	}

	return results, nil
}

// UpdateIndex sets data into a content type's search index at the given
// identifier
func (s *Search) UpdateIndex(ns, id string, data []byte) error {
	idx, ok := SearchIndices[ns]
	if ok {
		// unmarshal json to struct, error if not registered
		it, ok := ContentTypes[ns]
		if !ok {
			return fmt.Errorf("[search] UpdateIndex Error: type '%s' doesn't exist", ns)
		}

		p := it()
		err := json.Unmarshal(data, &p)
		if err != nil {
			return err
		}

		// add data to search index
		i := valueobject.NewIndex(ns, id)
		return idx.Index(i.String(), p)
	}

	return nil
}

// DeleteIndex removes data from a content type's search index at the
// given identifier
func (s *Search) DeleteIndex(id string) error {
	// check if there is a search index to work with
	target := strings.Split(id, ":")
	ns := target[0]

	idx, ok := SearchIndices[ns]
	if ok {
		// add data to search index
		return idx.Delete(id)
	}

	return nil
}

// Setup initializes Search Index for search to be functional
// This was moved out of db.Init and put to main(), because addon checker was initializing db together with
// search indexing initialisation in time when there were no item.Types defined so search index was always
// empty when using addons. We still have no guarentee whatsoever that item.Types is defined
// Should be called from a goroutine after SetContent is successful (SortContent requirement)
func (s *Search) Setup(cts map[string]content.Creator, searchDir string) {
	SearchIndices = make(map[string]bleve.Index)
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
	SearchIndices[typeName] = idx
	fmt.Printf("[search] Index created for %s\n", typeName)

	return nil
}

package entity

import (
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/internal/domain/content/repository"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"os"
	"path/filepath"
	"strings"
)

type TypeService interface {
	GetContentCreator(name string) (content.Creator, bool)
	AllContentTypeNames() []string
	AllAdminTypeNames() []string
	IsAdminType(name string) bool
}

type Search struct {
	TypeService TypeService

	Repo repository.Repository
	Log  loggers.Logger

	IndicesMap map[string]map[string]bleve.Index
}

// TypeQuery conducts a search and returns a set of Ponzu "targets", Type:ID pairs,
// and an error. If there is no search index for the typeName (Type) provided,
// db.ErrNoIndex will be returned as the error
func (s *Search) TypeQuery(typeName, query string, count, offset int) ([]content.Identifier, error) {
	s.setup()

	idx, ok := s.IndicesMap[s.getSearchDir(typeName)][typeName]
	if !ok {
		s.Log.Debugln("Index for type ", typeName, " not found")
		return nil, content.ErrNoIndex
	}

	s.Log.Debugln("TypeQuery: ", query)
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
	s.setup()

	idx, ok := s.IndicesMap[s.getSearchDir(ns)][ns]
	if ok {
		// unmarshal json to struct, error if not registered
		it, ok := s.TypeService.GetContentCreator(ns)
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
	s.setup()

	// check if there is a search index to work with
	target := strings.Split(id, ":")
	ns := target[0]

	idx, ok := s.IndicesMap[s.getSearchDir(ns)][ns]
	if ok {
		// add data to search index
		return idx.Delete(id)
	}

	return nil
}

func (s *Search) getSearchDir(ns string) string {
	if s.TypeService.IsAdminType(ns) {
		return s.adminSearchDir()
	}
	return s.userSearchDir()
}

func (s *Search) userSearchDir() string {
	return filepath.Join(s.Repo.UserDataDir(), "Search")
}

func (s *Search) adminSearchDir() string {
	return filepath.Join(s.Repo.AdminDataDir(), "Search")
}

// Setup initializes Search Index for search to be functional
// This was moved out of db.Init and put to main(), because addon checker was initializing db together with
// search indexing initialisation in time when there were no item.Types defined so search index was always
// empty when using addons. We still have no guarentee whatsoever that item.Types is defined
// Should be called from a goroutine after SetContent is successful (SortContent requirement)
func (s *Search) setup() {
	s.setupAdminIndices()
	s.setupUserIndices()
}

func (s *Search) setupUserIndices() {
	s.setupIndices(s.userSearchDir(), s.TypeService.AllContentTypeNames())
}

func (s *Search) setupAdminIndices() {
	s.setupIndices(s.adminSearchDir(), s.TypeService.AllAdminTypeNames())
}

func (s *Search) setupIndices(dir string, typeNames []string) {
	_, ok := s.IndicesMap[dir]
	if ok {
		return
	}

	searchIndices := make(map[string]bleve.Index)

	for _, t := range typeNames {
		idx, err := s.mapIndex(t)
		if err != nil {
			s.Log.Errorln("[search] Setup Error", err)
			return
		}

		searchIndices[t] = idx
	}

	s.IndicesMap[dir] = searchIndices
}

// MapIndex creates the mapping for a type and tracks the index to be used within
// the system for adding/deleting/checking data
func (s *Search) mapIndex(typeName string) (bleve.Index, error) {
	it, ok := s.TypeService.GetContentCreator(typeName)
	if !ok {
		return nil, fmt.Errorf("[search] MapIndex Error: Failed to MapIndex for %s, type doesn't exist", typeName)
	}
	sc, ok := it().(content.Searchable)
	if !ok {
		return nil, fmt.Errorf("[search] MapIndex Error: Item type %s doesn't implement search.Searchable", typeName)
	}

	// skip setting or using index for types that shouldn't be indexed
	if !sc.IndexContent() {
		s.Log.Warnf("[search] Index not created for %s\n", typeName)
		return nil, nil
	}

	mapping, err := sc.SearchMapping()
	if err != nil {
		return nil, err
	}

	idxName := typeName + ".index"
	var idx bleve.Index

	searchPath := s.getSearchDir(typeName)

	err = os.MkdirAll(searchPath, os.ModeDir|os.ModePerm)
	if err != nil {
		return nil, err
	}

	idxPath := filepath.Join(searchPath, idxName)
	if _, err = os.Stat(idxPath); os.IsNotExist(err) {
		idx, err = bleve.New(idxPath, mapping)
		if err != nil {
			return nil, err
		}
		idx.SetName(idxName)
	} else {
		idx, err = bleve.Open(idxPath)
		if err != nil {
			return nil, err
		}
	}

	s.Log.Debugf("[search] Index created for %s\n", typeName)

	return idx, nil
}

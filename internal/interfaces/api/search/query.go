package search

import (
	"fmt"
	"github.com/blevesearch/bleve"
)

// TypeQuery conducts a search and returns a set of Ponzu "targets", Type:ID pairs,
// and an error. If there is no search index for the typeName (Type) provided,
// db.ErrNoIndex will be returned as the error
func TypeQuery(typeName, query string, count, offset int) ([]Index, error) {
	idx, ok := Search[typeName]
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

	var results []Index
	for _, hit := range res.Hits {
		results = append(results, CreateIndex(hit.ID))
	}

	return results, nil
}

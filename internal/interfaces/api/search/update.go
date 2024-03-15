package search

import (
	"encoding/json"
	"fmt"
)

// UpdateIndex sets data into a content type's search index at the given
// identifier
func UpdateIndex(ns, id string, data []byte) error {
	idx, ok := Search[ns]
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
		i := &index{ns: ns, id: id}
		return idx.Index(i.String(), p)
	}

	return nil
}

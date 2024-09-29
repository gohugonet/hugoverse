package valueobject

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Index struct {
	ns string
	id string
}

func (i *Index) String() string {
	return fmt.Sprintf("%s:%s", i.ns, i.id)
}

func (i *Index) Namespace() string {
	return i.ns
}

func (i *Index) ContentType() string {
	return i.ns
}

func (i *Index) ID() string {
	return i.id
}

func CreateIndex(target string) *Index {
	t := strings.Split(target, ":")
	return &Index{
		ns: t[0],
		id: t[1],
	}
}

func NewIndex(ns, id string) *Index {
	return &Index{
		ns: ns,
		id: id,
	}
}

func GetIdFromQueryString(queryString string) (string, error) {
	// Parse the query string
	values, err := url.ParseQuery(queryString)
	if err != nil {
		return "", fmt.Errorf("failed to parse query string: %v", err)
	}

	// Extract the 'id' parameter
	id := values.Get("id")
	if id == "" {
		return "", errors.New("id parameter not found")
	}

	return id, nil
}

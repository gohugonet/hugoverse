package api

import "github.com/gohugonet/hugoverse/pkg/db"

type database struct {
}

func (d *database) start(contentTypeNames []string) {
	db.Start(dataDir(), contentTypeNames)
}

func (d *database) close() {
	db.Close()
}

func (d *database) PutConfig(key string, value interface{}) error {
	//todo, abstract to Put

	return nil
}

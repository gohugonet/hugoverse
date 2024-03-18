package factory

import (
	"github.com/gofrs/uuid"
	"github.com/gohugonet/hugoverse/internal/domain/content/valueobject"
	"github.com/gohugonet/hugoverse/pkg/timestamp"
)

func NewItem() (*valueobject.Item, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	nowMillis := timestamp.CurrentTimeMillis()

	return &valueobject.Item{
		UUID:      uid,
		ID:        -1,
		Slug:      "",
		Timestamp: nowMillis,
		Updated:   nowMillis,
	}, nil
}

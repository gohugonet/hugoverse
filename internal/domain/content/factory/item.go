package factory

import (
	"github.com/gofrs/uuid"
	"github.com/gohugonet/hugoverse/internal/domain/content/entity"
	"github.com/gohugonet/hugoverse/pkg/timestamp"
)

func NewItem() (*entity.Item, error) {
	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	nowMillis := timestamp.CurrentTimeMillis()

	return &entity.Item{
		UUID:      uid,
		ID:        -1,
		Slug:      "",
		Timestamp: nowMillis,
		Updated:   nowMillis,
	}, nil
}

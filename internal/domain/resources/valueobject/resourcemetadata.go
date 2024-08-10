package valueobject

import (
	"encoding/json"
)

type ResourceMetadata struct {
	Target     string         `json:"Target"`
	MediaTypeV string         `json:"MediaType"`
	MetaData   map[string]any `json:"Data"`
}

func (r ResourceMetadata) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r ResourceMetadata) Unmarshal(data []byte) (ResourceMetadata, error) {
	var r2 ResourceMetadata
	err := json.Unmarshal(data, &r2)
	return r2, err
}

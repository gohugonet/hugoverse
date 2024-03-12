package entity

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/gofrs/uuid"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"net/http"
)

// Item should only be embedded into content type structs.
type Item struct {
	UUID      uuid.UUID      `json:"uuid"`
	Status    content.Status `json:"status"`
	Name      string         `json:"name"`
	ID        int            `json:"id"`
	Slug      string         `json:"slug"`
	Timestamp int64          `json:"timestamp"`
	Updated   int64          `json:"updated"`
}

// Time partially implements the Sortable interface
func (i *Item) Time() int64 {
	return i.Timestamp
}

// Touch partially implements the Sortable interface
func (i *Item) Touch() int64 {
	return i.Updated
}

// SetSlug sets the item's slug for its URL
func (i *Item) SetSlug(slug string) {
	i.Slug = slug
}

// ItemSlug sets the item's slug for its URL
func (i *Item) ItemSlug() string {
	return i.Slug
}

// ItemID gets the *Item's ID field
// partially implements the Identifiable interface
func (i *Item) ItemID() int {
	return i.ID
}

// ItemName gets the *Item's Name field
// partially implements the Identifiable interface
func (i *Item) ItemName() string {
	return i.Name
}

// SetItemID sets the *Item's ID field
// partially implements the Identifiable interface
func (i *Item) SetItemID(id int) {
	i.ID = id
}

// UniqueID gets the *Item's UUID field
// partially implements the Identifiable interface
func (i *Item) UniqueID() uuid.UUID {
	return i.UUID
}

func (i *Item) SetUniqueID(uuid uuid.UUID) {
	i.UUID = uuid
}

// SetItemStatus sets the *Item's Status field
// partially implements the Identifiable interface
func (i *Item) SetItemStatus(status content.Status) {
	i.Status = status
}

// ItemStatus gets the *Item's Status field
// partially implements the Identifiable interface
func (i *Item) ItemStatus() content.Status {
	return i.Status
}

// String formats an *Item into a printable value
// partially implements the Identifiable interface
func (i *Item) String() string {
	return fmt.Sprintf("Item id: %s", i.UniqueID())
}

// BeforeAPIResponse is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeAPIResponse(res http.ResponseWriter, req *http.Request, data []byte) ([]byte, error) {
	return data, nil
}

// AfterAPIResponse is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterAPIResponse(res http.ResponseWriter, req *http.Request, data []byte) error {
	return nil
}

// BeforeAPICreate is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeAPICreate(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterAPICreate is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterAPICreate(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeAPIUpdate is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeAPIUpdate(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterAPIUpdate is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterAPIUpdate(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeAPIDelete is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeAPIDelete(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterAPIDelete is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterAPIDelete(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeAdminCreate is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeAdminCreate(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterAdminCreate is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterAdminCreate(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeAdminUpdate is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeAdminUpdate(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterAdminUpdate is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterAdminUpdate(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeAdminDelete is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeAdminDelete(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterAdminDelete is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterAdminDelete(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeSave is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeSave(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterSave is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterSave(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeDelete is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeDelete(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterDelete is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterDelete(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeApprove is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeApprove(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterApprove is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterApprove(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeReject is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeReject(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterReject is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterReject(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeEnable is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeEnable(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterEnable is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterEnable(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// BeforeDisable is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) BeforeDisable(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// AfterDisable is a no-op to ensure structs which embed *Item implement Hookable
func (i *Item) AfterDisable(res http.ResponseWriter, req *http.Request) error {
	return nil
}

// SearchMapping returns a default implementation of a Bleve IndexMappingImpl
// partially implements search.Searchable
func (i *Item) SearchMapping() (*mapping.IndexMappingImpl, error) {
	mapping := bleve.NewIndexMapping()
	mapping.StoreDynamic = false

	return mapping, nil
}

// IndexContent determines if a type should be indexed for searching
// partially implements search.Searchable
func (i *Item) IndexContent() bool {
	return false
}

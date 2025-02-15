package repository

type Repository interface {
	PutConfig(data []byte) error
	LoadConfig() ([]byte, error)

	User(email string) ([]byte, error)
	Users() [][]byte
	PutUser(email string, data []byte) error
	NextUserId(email string) (uint64, error)

	NewUpload(id, slug string, data []byte) error
	NextUploadId() (uint64, error)
	GetUpload(id string) ([]byte, error)
	DeleteUpload(id string) error
	AllUploads() ([][]byte, error)

	CheckSlugForDuplicate(namespace string, slug string) (string, error)
}

package database

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content"
	"github.com/gohugonet/hugoverse/pkg/db"
	"github.com/gohugonet/hugoverse/pkg/loggers"
	"path"
	"strconv"
)

var (
	adminBuckets = []string{
		"__config", "__users",
	}

	userBuckets = []string{
		"__addons", "__uploads",
		"__contentIndex",
	}
)

type Database struct {
	dataDir        string
	userDir        string
	contentBuckets []string

	adminStore *db.Store
	userStore  *db.Store

	log loggers.Logger
}

func New(dataDir string) (*Database, error) {
	as, err := db.NewStore(dataDir, adminBuckets)
	if err != nil {
		return nil, err
	}

	return &Database{
		dataDir: dataDir,

		adminStore: as,

		log: loggers.NewDefault(),
	}, nil
}

func (d *Database) UserDataDir() string {
	return path.Join(d.dataDir, d.userDir)
}

func (d *Database) RegisterContentBuckets(contentTypeNames []string) {
	d.contentBuckets = append(d.contentBuckets, contentTypeNames...)
}

func (d *Database) Close() {
	if d.userStore != nil {
		d.userStore.Close()
	}
	if d.adminStore != nil {
		d.adminStore.Close()
	}
}

func (d *Database) PutSortedContent(namespace string, m map[string][]byte) error {
	return d.userStore.Sort(newItems(namespace, m))
}

func (d *Database) AllContent(namespace string) [][]byte {
	return d.userStore.ContentAll(namespace)
}

func (d *Database) GetContent(contentType string, id string) ([]byte, error) {
	return d.userStore.Get(
		&item{
			bucket: contentType,
			key:    id,
		})
}

func (d *Database) DeleteContent(namespace string, id string, slug string) error {
	if err := d.userStore.Delete(&item{bucket: namespace, key: id}); err != nil {
		return err
	}

	if err := d.userStore.RemoveIndex(slug); err != nil {
		return err
	}

	return nil
}

func (d *Database) PutContent(ci any, data []byte) error {
	cii, ok := ci.(content.Identifiable)
	if !ok {
		return errors.New("invalid content type")
	}
	id := cii.ItemID()
	ns := cii.ItemName()

	cis, ok := ci.(content.Statusable)
	if !ok {
		return errors.New("invalid content type")
	}
	status := cis.ItemStatus()

	bucket := ns
	if !(status == content.Public || status == "") {
		bucket = fmt.Sprintf("%s%s", ns, bucketNameWithPrefix(string(status)))
	}

	fmt.Printf(" === bucket: %s\n", bucket)

	if err := d.userStore.Set(
		&item{
			bucket: bucket,
			key:    strconv.FormatInt(int64(id), 10),
			value:  data,
		}); err != nil {
		return err
	}

	return nil
}

func (d *Database) NewContent(ci any, data []byte) error {
	if err := d.PutContent(ci, data); err != nil {
		return err
	}

	cii, ok := ci.(content.Identifiable)
	if !ok {
		return errors.New("invalid content type")
	}
	id := cii.ItemID()
	ns := cii.ItemName()

	ciSlug, ok := ci.(content.Sluggable)
	if !ok {
		return errors.New("invalid content type")
	}
	if err := d.userStore.SetIndex(newKeyValueItem(ciSlug.ItemSlug(), fmt.Sprintf("%s:%d", ns, id))); err != nil {
		return err
	}

	return nil
}

func (d *Database) NextContentId(ns string) (uint64, error) {
	return d.userStore.NextSequence(&item{bucket: ns})
}

func (d *Database) NextUploadId() (uint64, error) {
	return d.userStore.NextSequence(newBucketItem("uploads"))
}

func (d *Database) GetUpload(id string) ([]byte, error) {
	key, err := keyBit8Uint64(id)
	if err != nil {
		return nil, err
	}
	return d.userStore.Get(newUploadItem(key, nil))
}

func (d *Database) DeleteUpload(id string) error {
	key, err := keyBit8Uint64(id)
	if err != nil {
		return err
	}
	return d.userStore.Delete(newUploadItem(key, nil))
}

func (d *Database) AllUploads() ([][]byte, error) {
	return db.All(newUploadItem("", nil))
}

func (d *Database) NewUpload(id, slug string, data []byte) error {
	key, err := keyBit8Uint64(id)
	if err != nil {
		return err
	}
	if err := d.userStore.Set(newUploadItem(key, data)); err != nil {
		return err
	}

	if err := d.userStore.SetIndex(newKeyValueItem(
		slug,
		fmt.Sprintf("%s:%s", bucketNameWithPrefix("upload"), id))); err != nil {
		return err
	}

	return nil
}

func keyBit8Uint64(sid string) (string, error) {
	id, err := strconv.Atoi(sid)
	if err != nil {
		return "", err
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return string(b), err
}

func (d *Database) CheckSlugForDuplicate(slug string) (string, error) {
	return d.userStore.CheckSlugForDuplicate(slug)
}

func (d *Database) SystemInitComplete() bool {
	users := d.adminStore.ContentAll(bucketNameWithPrefix("users"))

	return len(users) > 0
}

func (d *Database) Query(namespace string, opts db.QueryOptions) (int, [][]byte) {
	return d.userStore.Query(namespace, opts)
}

package database

import "github.com/mdfriday/hugoverse/pkg/db"

func (d *Database) StartAdminDatabase(adminTypeNames []string) error {
	var buckets []string
	buckets = append(buckets, adminTypeNames...)
	buckets = append(buckets, adminOriginBuckets...)

	as, err := db.NewStore(d.dataDir, buckets)
	if err != nil {
		return err
	}

	d.adminStore = as
	d.adminBuckets = adminTypeNames

	return nil
}

func (d *Database) SystemInitComplete() bool {
	users := d.adminStore.ContentAll(bucketNameWithPrefix("users"))

	return len(users) > 0
}

func (d *Database) User(email string) ([]byte, error) {
	return d.adminStore.Get(newUserItem(email, nil))
}

func (d *Database) Users() [][]byte {
	return d.adminStore.ContentAll(bucketNameWithPrefix("users"))
}

func (d *Database) PutUser(email string, data []byte) error {
	return d.adminStore.Set(newUserItem(email, data))
}

func (d *Database) NextUserId(email string) (uint64, error) {
	return d.adminStore.NextSequence(newBucketItem("users"))
}

func (d *Database) PutConfig(data []byte) error {
	return d.adminStore.Set(newConfigItem(data))
}

func (d *Database) LoadConfig() ([]byte, error) {
	data, err := d.adminStore.Get(newConfigItem(nil))
	if err != nil {
		return nil, err
	}

	return data, nil
}

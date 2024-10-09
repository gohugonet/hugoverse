package database

func (d *Database) User(email string) ([]byte, error) {
	return d.adminStore.Get(newUserItem(email, nil))
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

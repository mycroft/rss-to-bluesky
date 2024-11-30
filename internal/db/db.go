package db

import bolt "go.etcd.io/bbolt"

type DB struct {
	db *bolt.DB
}

func Open() (*DB, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Set(key string, value []byte) error {
	return set(d.db, key, value)
}

func (d *DB) Get(key string) ([]byte, error) {
	return get(d.db, key)
}

func (d *DB) Has(key string) (bool, error) {
	return has(d.db, key)
}

func openDB() (*bolt.DB, error) {
	db, err := bolt.Open("./database/db", 0666, nil)
	if err != nil {
		return nil, err
	}

	return db, err
}

func get(db *bolt.DB, key string) ([]byte, error) {
	var value []byte

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("records"))
		if b == nil {
			return nil
		}

		value = b.Get([]byte(key))

		return nil
	})

	return value, err
}

func set(db *bolt.DB, key string, value []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("records"))
		if err != nil {
			return err
		}

		return b.Put([]byte(key), []byte(value))
	})

	return err
}

func has(db *bolt.DB, key string) (bool, error) {
	found := false
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("records"))
		if b == nil {
			return nil
		}

		v := b.Get([]byte(key))
		if v == nil {
			return nil
		}

		found = true

		return nil
	})

	return found, err
}

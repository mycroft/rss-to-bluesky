package db

import bolt "go.etcd.io/bbolt"

func OpenDB() (*bolt.DB, error) {
	db, err := bolt.Open("./db", 0666, nil)
	if err != nil {
		return nil, err
	}

	return db, err
}

func Set(db *bolt.DB, key, value string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("records"))
		if err != nil {
			return err
		}

		return b.Put([]byte(key), []byte("1"))
	})

	return err
}

func Has(db *bolt.DB, key string) (bool, error) {
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

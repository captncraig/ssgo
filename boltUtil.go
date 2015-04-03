package ssgo

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

func storeJson(db *bolt.DB, bucket string, key string, data interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(key), j)
		return err
	})
}

func lookupJson(db *bolt.DB, bucket string, key string, v interface{}) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("Not found")
		}
		return json.Unmarshal(data, v)
	})
}

func ensureBucketExists(db *bolt.DB, bucket string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

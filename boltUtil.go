package ssgo

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

func init() {
	boltPath := os.Getenv("ssgo.boltdb")
	if boltPath == "" {
		boltPath = "ssgo.db"
	}
	var err error
	db, err = bolt.Open(boltPath, 0600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		panic(err.Error())
	}
}

func StoreBoltJson(bucket string, key string, data interface{}) error {
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

func LookupBoltJson(bucket string, key string, v interface{}) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("Not found")
		}
		return json.Unmarshal(data, v)
	})
}

func EnsureBoltBucketExists(bucket string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

package boltstore

import (
	"errors"
	"github.com/boltdb/bolt"
	"github.com/dgruber/drmaa2os/pkg/storage"
	"log"
	"time"
)

func NewBoltStore(path string) storage.Storer {
	return &BoltStore{dbfile: path}
}

type BoltStore struct {
	dbfile string
	db     *bolt.DB
}

func (b *BoltStore) Init() error {
	var err error
	b.db, err = bolt.Open(b.dbfile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (b *BoltStore) Exit() error {
	if b.db != nil {
		return b.db.Close()
	}
	return errors.New("No DB handle")
}

func (b *BoltStore) Put(t storage.KeyType, key, value string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(t.String()))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(value))
	})
}

func (b *BoltStore) Get(t storage.KeyType, key string) (string, error) {
	var data []byte
	var err error
	b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t.String()))
		if b == nil {
			data = make([]byte, 0)
			err = errors.New("Not found!")
			return nil
		}
		r := b.Get([]byte(key))
		if r != nil {
			data = make([]byte, len(r))
			copy(data, r)
			return nil
		}
		data = make([]byte, 0)
		err = errors.New("Not found!")
		return nil
	})
	return string(data), err
}

func (b *BoltStore) List(t storage.KeyType) ([]string, error) {
	keys := make([]string, 0, 1024)
	b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t.String()))
		if b == nil {
			// no list defined
			return nil
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			keys = append(keys, string(k))
		}
		return nil
	})
	return keys, nil
}

func (b *BoltStore) Delete(t storage.KeyType, key string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(t.String()))
		if b == nil {
			return errors.New("it does not exist")
		}
		if b.Get([]byte(key)) == nil {
			return errors.New("it does not exist")
		}
		return b.Delete([]byte(key))
	})
}

func (b *BoltStore) Exists(t storage.KeyType, key string) bool {
	if value, err := b.Get(t, key); err == nil {
		if value != "" {
			return true
		}
	}
	return false
}

package persistence

import (
	"context"
	"encoding/json"
	"github.com/caarlos0/log"
	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"sync"
	"time"
)

var (
	KeyList = []byte("keyList")
	wg      = &sync.WaitGroup{}
)

type (
	Persistence interface {
		// Get returns the value for the given key.
		Get(key []byte) ([]byte, error)

		// Set sets the value for the given key.
		Set(key, value []byte) error

		// Update updates the value for the given key.
		Update(key, value []byte) error

		// Delete deletes the value for the given key.
		Delete(key []byte) error

		// Close closes the database and frees up any resources.
		Close() error

		// List returns a list of all keyList in the key-value store.
		List() (map[string]KeyValue, error)
	}

	// BadgerPersistence is a wrapper around BadgerDB.
	BadgerPersistence struct {
		ctx     context.Context
		db      *badger.DB
		keyList map[string]KeyValue
		addKey  chan KeyValue
		mutex   *sync.Mutex
	}

	KeyValue struct {
		Name     string
		Key      []byte
		CreateAt int64
	}
)

func NewBadgerPersistence(ctx context.Context, path string) (*BadgerPersistence, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	b := &BadgerPersistence{
		db:      db,
		keyList: make(map[string]KeyValue),
		addKey:  make(chan KeyValue, 10),
		mutex:   &sync.Mutex{},
		ctx:     ctx,
	}

	go b.keyAppender()

	return b, nil
}

//func (b *BadgerPersistence) toHexValue(key []byte) string {
//	return fmt.Sprintf("%x", key)
//}

func (b *BadgerPersistence) keyAppender() {
	wg.Add(1)
	defer wg.Done()

	keyList, err := b.List()
	if err != nil {
		return
	}

	for key, value := range keyList {
		b.keyList[key] = value
	}

	for {
		select {
		case key := <-b.addKey:
			b.keyList[key.Name] = key
		case <-b.ctx.Done():
			return
		}
	}
}

func (b *BadgerPersistence) composeKey(uuidKey string) KeyValue {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var pk = uuid.MustParse(uuidKey)

	return KeyValue{
		Key:      pk.NodeID(),
		Name:     pk.String(),
		CreateAt: time.Now().UnixNano(),
	}
}

func (b *BadgerPersistence) appendKey(value KeyValue) error {
	data, err := b.serialize(value)
	if err != nil {
		return err
	}

	if err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(KeyList, data)
	}); err != nil {
		return err
	}

	return nil
}

func (b *BadgerPersistence) serialize(value any) ([]byte, error) {
	return json.Marshal(value)
}

func (b *BadgerPersistence) Get(key string) ([]byte, error) {
	if k, ok := b.keyList[key]; ok {
		var value []byte
		err := b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(k.Key)
			if err != nil {
				return err
			}

			value, err = item.ValueCopy(nil)
			return err
		})

		return value, err
	}

	return nil, nil
}

func (b *BadgerPersistence) Set(value []byte) (string, error) {
	return b.SetKey(uuid.NewString(), value)
}

func (b *BadgerPersistence) SetKey(uuidKey string, value []byte) (string, error) {
	genKey := b.composeKey(uuidKey)
	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(genKey.Key, value)
	}); err != nil {
		return "", err
	}

	b.addKey <- genKey

	return genKey.Name, nil
}

func (b *BadgerPersistence) Delete(key string) error {
	if k, ok := b.keyList[key]; ok {
		if err := b.db.Update(func(txn *badger.Txn) error {
			return txn.Delete(k.Key)
		}); err != nil {
			return err
		}

		delete(b.keyList, key)
	}

	return nil
}

func (b *BadgerPersistence) Close() {
	if err := b.db.Close(); err != nil {
		log.Fatalf("Error closing Badger database:", err)
	}
}

func (b *BadgerPersistence) List() (map[string]KeyValue, error) {
	var keys = make(map[string]KeyValue)
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			pk := item.KeyCopy(nil)
			id, err := uuid.FromBytes(pk)
			if err != nil {
				return err
			}
			keys[id.String()] = KeyValue{
				Key:      uuid.New().NodeID(),
				CreateAt: time.Now().UnixNano(),
			}
		}

		return nil
	})

	return keys, err
}

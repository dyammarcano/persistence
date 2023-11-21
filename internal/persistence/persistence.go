package persistence

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/caarlos0/log"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
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
		List() (map[string][]byte, error)
	}

	// BadgerPersistence is a wrapper around BadgerDB.
	BadgerPersistence struct {
		ctx     context.Context
		db      *badger.DB
		keyList map[string][]byte
		addKey  chan []byte
		mutex   *sync.Mutex
	}

	Key struct {
		Key      []byte
		TTL      int64
		Hits     int
		CreateAt int64
		UpdateAt int64
		Data     any
	}

	Value struct {
		CreateAt int64
		UpdateAt int64
		Payload  []byte
	}
)

func NewBadgerPersistence(ctx context.Context, path string) (*BadgerPersistence, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	b := &BadgerPersistence{
		db:      db,
		keyList: make(map[string][]byte),
		addKey:  make(chan []byte, 10),
		mutex:   &sync.Mutex{},
		ctx:     ctx,
	}

	go b.keysMonitor()

	if err = b.loadKeys(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *BadgerPersistence) keysMonitor() {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case key := <-b.addKey:
			b.keyList[b.encodeKey(key)] = key
		case <-b.ctx.Done():
			return
		}
	}
}

func (b *BadgerPersistence) encodeKey(key []byte) string {
	return hex.EncodeToString(key)
}

func (b *BadgerPersistence) decodeKey(key string) ([]byte, error) {
	return hex.DecodeString(key)
}

func (b *BadgerPersistence) composeKey(key string) *Key {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	pk, err := ulid.Parse(key)
	if err != nil {
		return nil
	}

	return &Key{
		Key:      pk.Bytes(),
		CreateAt: time.Now().UnixNano(),
		UpdateAt: time.Now().UnixNano(),
	}
}

func (b *BadgerPersistence) appendKey(value Key) error {
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

func (b *BadgerPersistence) deserialize(val []byte) (*Value, error) {
	var obj Value
	err := json.Unmarshal(val, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (b *BadgerPersistence) Get(key []byte) (*Value, error) {
	if k, ok := b.keyList[string(key)]; ok {
		var value *Value
		err := b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(k)
			if err != nil {
				return err
			}

			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			value, err = b.deserialize(val)
			if err != nil {
				return err
			}

			return err
		})

		return value, err
	}

	return nil, nil
}

func (b *BadgerPersistence) Set(value []byte) (*Key, error) {
	return b.SetWithKey(ulid.Make().String(), value)
}

func (b *BadgerPersistence) SetWithKey(key string, value []byte) (*Key, error) {
	genKey := b.composeKey(key)

	data, err := b.serialize(Value{
		CreateAt: genKey.CreateAt,
		UpdateAt: genKey.UpdateAt,
		Payload:  value,
	})
	if err != nil {
		return nil, err
	}

	if err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(genKey.Key, data)
	}); err != nil {
		return nil, err
	}

	b.addKey <- genKey.Key

	return genKey, nil
}

func (b *BadgerPersistence) Delete(key string) error {
	if k, ok := b.keyList[key]; ok {
		if err := b.db.Update(func(txn *badger.Txn) error {
			return txn.Delete(k)
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

//func (b *BadgerPersistence) List() (map[string]Key, error) {
//	var keys = make(map[string]Key)
//	err := b.db.View(func(txn *badger.Txn) error {
//		it := txn.NewIterator(badger.DefaultIteratorOptions)
//		defer it.Close()
//
//		for it.Rewind(); it.Valid(); it.Next() {
//			item := it.Item()
//			pk := item.KeyCopy(nil)
//			id, err := uuid.FromBytes(pk)
//			if err != nil {
//				return err
//			}
//			keys[id.String()] = Key{
//				Key:      uuid.New().NodeID(),
//				CreateAt: time.Now().UnixNano(),
//			}
//		}
//
//		return nil
//	})
//
//	return keys, err
//}

func (b *BadgerPersistence) loadKeys() error {
	return b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			b.addKey <- item.KeyCopy(nil)
		}

		return nil
	})
}

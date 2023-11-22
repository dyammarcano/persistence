package persistence

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/badger/v4"
	"github.com/oklog/ulid/v2"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
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
		db       *badger.DB
		keyList  map[string][]byte
		addKeyCh chan []byte
		mutex    *sync.Mutex
		wg       sync.WaitGroup
		ctx      context.Context
	}

	Key struct {
		String   string
		Key      []byte
		Hits     int
		CreateAt int64
		UpdateAt int64
	}

	Value struct {
		CreateAt int64
		UpdateAt int64
		Payload  []byte
	}
)

// NewBadgerPersistence returns a new BadgerPersistence.
func NewBadgerPersistence(ctx context.Context, path string) (*BadgerPersistence, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	b := &BadgerPersistence{
		db:       db,
		keyList:  make(map[string][]byte),
		addKeyCh: make(chan []byte, 10),
		mutex:    &sync.Mutex{},
		wg:       sync.WaitGroup{},
		ctx:      ctx,
	}

	go b.keysMonitor()

	if err = b.loadKeys(); err != nil {
		return nil, err
	}
	return b, nil
}

// keysMonitor monitors the addKeyCh channel.
func (b *BadgerPersistence) keysMonitor() {
	b.wg.Add(1)
	defer b.wg.Done()

	for {
		select {
		case key := <-b.addKeyCh:
			b.addKey(key)
		case <-b.ctx.Done():
			return
		}
	}
}

// loadKeys loads all keys from the database.
func (b *BadgerPersistence) loadKeys() error {
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			b.addKeyCh <- item.KeyCopy(nil)
		}
		it.Close()
		return nil
	})
	return err
}

// addKey adds a key to the keyList.
func (b *BadgerPersistence) addKey(key []byte) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.keyList[hex.EncodeToString(key)] = key
}

// getKey returns a key from the keyList.
func (b *BadgerPersistence) getKey(key string) ([]byte, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if k, ok := b.keyList[key]; ok {
		return k, nil
	}
	return nil, ErrKeyNotFound
}

// composeKey generate a new key.
func (b *BadgerPersistence) composeKey(key string) *Key {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	id, _ := ulid.Parse(key)
	return &Key{
		Key:      id.Bytes(),
		String:   hex.EncodeToString(id.Bytes()),
		CreateAt: time.Now().UnixNano(),
		UpdateAt: time.Now().UnixNano(),
	}
}

// serialize serializes a value.
func (b *BadgerPersistence) serialize(value any) ([]byte, error) {
	return json.Marshal(value)
}

// deserialize deserializes a value.
func (b *BadgerPersistence) deserialize(obj *Value, val []byte) error {
	if err := json.Unmarshal(val, &obj); err != nil {
		return err
	}
	return nil
}

// Get returns the value for the given key.
func (b *BadgerPersistence) Get(key string) (*Value, error) {
	k, err := b.getKey(key)
	if err != nil {
		return nil, err
	}

	valObj := &Value{}
	err = b.db.View(func(txn *badger.Txn) error {
		item, _ := txn.Get(k)
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		if err = b.deserialize(valObj, val); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return valObj, err
}

// Set sets the value for the given key.
func (b *BadgerPersistence) Set(value []byte) (*Key, error) {
	return b.SetWithKey(ulid.Make().String(), value)
}

// SetWithKey sets the value for the given key.
func (b *BadgerPersistence) SetWithKey(key string, value []byte) (*Key, error) {
	var err error
	k := b.composeKey(key)

	val, err := b.serialize(Value{
		CreateAt: k.CreateAt,
		UpdateAt: k.UpdateAt,
		Payload:  value,
	})
	if err != nil {
		return nil, err
	}

	err = b.db.Update(func(txn *badger.Txn) error {
		err = txn.Set(k.Key, val)
		return err
	})
	if err != nil {
		return nil, err
	}

	b.addKeyCh <- k.Key
	return k, nil
}

// Delete deletes the value for the given key.
func (b *BadgerPersistence) Delete(key string) error {
	k, err := b.getKey(key)
	if err != nil {
		return err
	}
	err = b.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(k)
		return err
	})
	if err != nil {
		return err
	}

	delete(b.keyList, key)
	return nil
}

// Close closes the database and frees up any resources.
func (b *BadgerPersistence) Close() error {
	return b.db.Close()
}

// ListKeys returns a list of all keyList in the key-value store.
func (b *BadgerPersistence) ListKeys() (map[string][]byte, error) {
	return b.keyList, nil
}

// Len returns the number of keyList in the key-value store.
func (b *BadgerPersistence) Len() int {
	l := len(b.keyList)
	return l + 1
}

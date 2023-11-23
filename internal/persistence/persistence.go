package persistence

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/dyammarcano/base58"
	"strings"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type (
	Callback func(key string, err error)

	CachePersistence struct {
		db       *badger.DB
		keyList  map[string][]byte
		addKeyCh chan []byte
		mutex    *sync.Mutex
		ctx      context.Context
		expires  time.Duration
	}

	Key struct {
		String string
		Key    []byte
	}
)

// NewBadgerPersistenceWithInMemory returns a new CachePersistence with in-memory database.
func NewBadgerPersistenceWithInMemory(ctx context.Context) (*CachePersistence, error) {
	opts := badger.DefaultOptions("").WithInMemory(true)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	b := &CachePersistence{
		db:       db,
		keyList:  make(map[string][]byte),
		addKeyCh: make(chan []byte, 10),
		mutex:    &sync.Mutex{},
		ctx:      ctx,
		expires:  36 * time.Hour,
	}

	go b.keysMonitor()

	if err = b.loadKeys(); err != nil {
		return nil, err
	}
	return b, nil
}

// NewBadgerPersistence returns a new CachePersistence.
func NewBadgerPersistence(ctx context.Context, path string) (*CachePersistence, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	b := &CachePersistence{
		db:       db,
		keyList:  make(map[string][]byte),
		addKeyCh: make(chan []byte, 10),
		mutex:    &sync.Mutex{},
		ctx:      ctx,
		expires:  36 * time.Hour,
	}

	go b.keysMonitor()

	if err = b.loadKeys(); err != nil {
		return nil, err
	}
	return b, nil
}

// keysMonitor monitors the addKeyCh channel.
func (b *CachePersistence) keysMonitor() {
	for {
		select {
		case key := <-b.addKeyCh:
			// check if key already exists
			if _, ok := b.keyList[b.encodeKey(key)]; ok {
				continue
			}
			b.addKeyToKeyList(key)
		case <-b.ctx.Done():
			return
		}
	}
}

// loadKeys loads all keys from the database.
func (b *CachePersistence) loadKeys() error {
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			// check if key is expired or deleted
			if item.IsDeletedOrExpired() {
				continue
			}
			b.addKeyCh <- item.KeyCopy(nil)
		}
		it.Close()
		return nil
	})
	return err
}

// addKeyToKeyList adds a key to the keyList.
func (b *CachePersistence) addKeyToKeyList(key []byte) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.keyList[b.encodeKey(key)] = key
}

// getKeyFromKeyList returns a key from the keyList.
func (b *CachePersistence) getKeyFromKeyList(key string) ([]byte, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if k, ok := b.keyList[key]; ok {
		return k, nil
	}
	return nil, ErrKeyNotFound
}

// composeKey generate a new key.
func (b *CachePersistence) composeKey(key []byte) *Key {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return &Key{
		Key:    key,
		String: b.encodeKey(key),
	}
}

// generateRandomKey generate random key with fix length of 10 bits
func (b *CachePersistence) generateRandomKey() []byte {
	var vk = make([]byte, 10)
	if _, err := rand.Read(vk); err != nil {
		return nil
	}
	return vk
}

// encodeKey encodes a key and return a string with 27 characters.
func (b *CachePersistence) encodeKey(key []byte) string {
	return strings.ToUpper(base58.StdEncoding.EncodeToString(sha1.New().Sum(key)))[0:27]
}

// Serialize serializes a value.
func (b *CachePersistence) Serialize(value any) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize deserializes a value.
func (b *CachePersistence) Deserialize(obj any, val []byte) error {
	if err := gob.NewDecoder(bytes.NewReader(val)).Decode(obj); err != nil {
		return err
	}
	return nil
}

// Delete deletes the value for the given key.
func (b *CachePersistence) Delete(key string) error {
	ks, err := b.getKeyFromKeyList(key)
	if err != nil {
		return err
	}

	if err = b.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(ks)
		return err
	}); err != nil {
		return err
	}
	delete(b.keyList, key)
	return nil
}

// DeleteAll deletes all keyList in the key-value store.
func (b *CachePersistence) DeleteAll() error {
	if err := b.db.DropAll(); err != nil {
		return fmt.Errorf("failed to delete all keys: %w", err)
	}
	b.keyList = make(map[string][]byte)
	return nil
}

// Close closes the database and frees up any resources.
func (b *CachePersistence) Close() error {
	return b.db.Close()
}

// ListKeys returns a list of all keyList in the key-value store.
func (b *CachePersistence) ListKeys() map[string][]byte {
	return b.keyList
}

// Len returns the number of keyList in the key-value store.
func (b *CachePersistence) Len() int {
	l := len(b.keyList)
	return l + 1
}

// SetValue sets the value for the given key.
func (b *CachePersistence) SetValue(value []byte) (string, error) {
	vk := b.generateRandomKey()
	return b.SetValueWithKey(vk, value)
}

// SetValueWithKey sets the value for the given key.
func (b *CachePersistence) SetValueWithKey(key, value []byte) (string, error) {
	ks := b.composeKey(key)

	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(ks.Key, value).WithTTL(b.expires).WithMeta(byte(1)))
	}); err != nil {
		return "", err
	}
	b.addKeyCh <- ks.Key
	return ks.String, nil
}

// SetStruct sets the value for the given key.
func (b *CachePersistence) SetStruct(value any) (string, error) {
	data, err := b.Serialize(value)
	if err != nil {
		return "", err
	}
	return b.SetValue(data)
}

// SetStructAsync sets the value for the given key asynchronously.
func (b *CachePersistence) SetStructAsync(value any, fn Callback) {
	go fn(b.SetStruct(value))
}

// GetValue returns the value for the given key.
func (b *CachePersistence) GetValue(key string) ([]byte, error) {
	ks, err := b.getKeyFromKeyList(key)
	if err != nil {
		return nil, err
	}

	var result []byte
	if err = b.db.View(func(txn *badger.Txn) error {
		item, _ := txn.Get(ks)
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		result = val
		return nil
	}); err != nil {
		return nil, err
	}
	return result, err
}

// GetStruct returns the value for the given key.
func (b *CachePersistence) GetStruct(key string, value any) error {
	data, err := b.GetValue(key)
	if err != nil {
		return err
	}
	return b.Deserialize(value, data)
}

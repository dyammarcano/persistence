package persistence

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/dyammarcano/base58"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type (
	Callback func(key string, err error)

	// CachePersistence is a wrapper around BadgerDB.
	CachePersistence struct {
		db       *badger.DB
		keyList  map[string][]byte
		addKeyCh chan []byte
		mutex    *sync.Mutex
		wg       sync.WaitGroup
		ctx      context.Context
		expires  time.Duration
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
		wg:       sync.WaitGroup{},
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
	b.wg.Add(1)
	defer b.wg.Done()

	for {
		select {
		case key := <-b.addKeyCh:
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
		Key:      key,
		String:   b.encodeKey(key),
		CreateAt: time.Now().UnixNano(),
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

// serialize serializes a value.
func (b *CachePersistence) serialize(value any) ([]byte, error) {
	return json.Marshal(value)
}

// deserialize deserializes a value.
func (b *CachePersistence) deserialize(obj *Value, val []byte) error {
	if err := json.Unmarshal(val, &obj); err != nil {
		return err
	}
	return nil
}

// GetValue returns the value for the given key.
func (b *CachePersistence) GetValue(key string) (*Value, error) {
	ks, err := b.getKeyFromKeyList(key)
	if err != nil {
		return nil, err
	}

	valObj := &Value{}
	if err = b.db.View(func(txn *badger.Txn) error {
		item, _ := txn.Get(ks)
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		if err = b.deserialize(valObj, val); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return valObj, err
}

// SetValue sets the value for the given key.
func (b *CachePersistence) SetValue(value []byte) (string, error) {
	vk := b.generateRandomKey()
	return b.SetValueWithKey(vk, value)
}

// SetValueWithKey sets the value for the given key.
func (b *CachePersistence) SetValueWithKey(key, value []byte) (string, error) {
	ks := b.composeKey(key)

	val, err := b.serialize(Value{
		CreateAt: ks.CreateAt,
		Payload:  value,
	})
	if err != nil {
		return "", err
	}

	if err = b.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(badger.NewEntry(ks.Key, val).WithTTL(b.expires).WithMeta(byte(1)))
	}); err != nil {
		return "", err
	}
	b.addKeyCh <- ks.Key
	return ks.String, nil
}

// SetStruct sets the value for the given key.
func (b *CachePersistence) SetStruct(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return b.SetValue(data)
}

// SetStructAsync sets the value for the given key asynchronously.
func (b *CachePersistence) SetStructAsync(value any, fn Callback) {
	go fn(b.SetStruct(value))
}

// GetStruct returns the value for the given key.
func (b *CachePersistence) GetStruct(key string, value any) error {
	val, err := b.GetValue(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(val.Payload, &value)
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
func (b *CachePersistence) ListKeys() (map[string][]byte, error) {
	return b.keyList, nil
}

// Len returns the number of keyList in the key-value store.
func (b *CachePersistence) Len() int {
	l := len(b.keyList)
	return l + 1
}

// RegisterPersistenceWebInterface starts a web interface.
func (b *CachePersistence) RegisterPersistenceWebInterface(router *mux.Router) {
	router.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		keys, err := b.ListKeys()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		items := make(map[string]*Value, len(keys))
		for key := range keys {
			val, err := b.GetValue(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			items[key] = val
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(items); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.HandleFunc("/count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]int{"count": b.Len()}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	router.HandleFunc("/item/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["key"]

		val, err := b.GetValue(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(val); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//print persistent info like number of keys, etc
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]int{"Number of keys": b.Len()}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

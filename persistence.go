package persistence

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"log/slog"
	"sync"
	"time"
)

var (
	mutex = &sync.Mutex{}
)

type (
	Callback      func(key string, err error)
	processItem   func(item *badger.Item) error
	performAction func(txn *badger.Txn) error

	Store struct {
		db       *badger.DB
		keyList  map[string][]byte
		addKeyCh chan []byte
		ctx      context.Context
		expires  time.Duration
		logIdx   uint64
	}

	Key struct {
		String string
		Key    []byte
	}
)

// NewBadgerPersistence returns a new PersistenceStore.
func NewBadgerPersistence(ctx context.Context, memory bool, path string) (*Store, error) {
	opts := badger.DefaultOptions(path)
	if memory {
		opts = badger.DefaultOptions("").WithInMemory(true)
	}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	b := &Store{
		db:       db,
		keyList:  make(map[string][]byte),
		addKeyCh: make(chan []byte, 10),
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
func (s *Store) keysMonitor() {
	slog.Info("keysMonitor")
	for {
		select {
		case key := <-s.addKeyCh:
			// check if key already exists
			if _, ok := s.keyList[EncodeKey(key)]; ok {
				continue
			}
			s.addKeyToKeyList(key)
		case <-s.ctx.Done():
			return
		}
	}
}

// loadKeys loads all keys from the database.
func (s *Store) loadKeys() error {
	callback := func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		for it.Rewind(); it.ValidForPrefix(dbDatPrefix); it.Next() {
			item := it.Item()
			// check if key is expired or deleted
			if item.IsDeletedOrExpired() {
				continue
			}
			s.addKeyCh <- item.KeyCopy(nil)
		}
		it.Close()
		return nil
	}

	if err := s.badgerInteract("loadKeys", callback, true, false); err != nil {
		return err
	}
	return nil
}

// iterateDB iterates over all keyList in the key-value store.
func (s *Store) iterateDB(process processItem) error {
	return s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			if err := process(item); err != nil {
				return err
			}
		}
		return nil
	})
}

// addKeyToKeyList adds a key to the keyList.
func (s *Store) addKeyToKeyList(key []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	s.keyList[EncodeKey(key)] = key
}

// getKeyFromKeyList returns a key from the keyList.
func (s *Store) getKeyFromKeyList(key string) ([]byte, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if k, ok := s.keyList[key]; ok {
		return k, nil
	}
	return nil, errors.New("key not found")
}

// Delete deletes the value for the given key.
func (s *Store) Delete(key string) error {
	ks, err := s.getKeyFromKeyList(key)
	if err != nil {
		return err
	}

	callback := func(txn *badger.Txn) error {
		err := txn.Delete(ks)
		return err
	}

	if err = s.badgerInteract("Delete", callback, false, true); err != nil {
		return err
	}
	delete(s.keyList, key)
	return nil
}

// DropAll deletes all keyList in the key-value store.
func (s *Store) DropAll() error {
	slog.Info(fmt.Sprintf("DropAll, keyList: %d", len(s.keyList)))
	if err := s.db.DropAll(); err != nil {
		return fmt.Errorf("failed to delete all keys: %w", err)
	}
	s.keyList = make(map[string][]byte)
	return nil
}

// Close closes the database and frees up any resources.
func (s *Store) Close() error {
	return s.db.Close()
}

// ListKeys returns a list of all keyList in the key-value store.
func (s *Store) ListKeys() map[string][]byte {
	return s.keyList
}

// Length returns the number of keyList in the key-value store.
func (s *Store) Length() int {
	l := len(s.keyList)
	return l + 1
}

// Size get size of tables
func (s *Store) Size() int {
	return len(s.db.Tables())
}

// SetValue sets the value for the given key.
func (s *Store) SetValue(value []byte) (string, error) {
	vk := generateRandomKey()
	return s.SetValueWithKey(vk, value)
}

// SetValueWithKey sets the value for the given key.
func (s *Store) SetValueWithKey(key, value []byte) (string, error) {
	var keyStr string
	callback := func(txn *badger.Txn) error {
		ks := composeKey(key)
		err := txn.SetEntry(badger.NewEntry(ks.Key, value).WithTTL(s.expires).WithMeta(byte(1)))
		if err != nil {
			return err
		}
		s.addKeyCh <- ks.Key
		keyStr = ks.String
		return nil
	}

	if err := s.badgerInteract("SetValueWithKey", callback, false, true); err != nil {
		return "", err
	}
	return keyStr, nil
}

// SetStruct sets the value for the given key.
func (s *Store) SetStruct(value any) (string, error) {
	data, err := Serialize(value)
	if err != nil {
		return "", err
	}
	return s.SetValue(data)
}

// GetValue returns the value for the given key.
func (s *Store) GetValue(key string) ([]byte, error) {
	ks, err := s.getKeyFromKeyList(key)
	if err != nil {
		return nil, err
	}

	var result []byte
	callback := func(txn *badger.Txn) error {
		item, err := txn.Get(ks)
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		result = val
		return nil
	}

	if err = s.badgerInteract("GetValue", callback, true, false); err != nil {
		return nil, err
	}
	return result, err
}

// badgerInteract interacts with the badger database.
func (s *Store) badgerInteract(caller string, callback performAction, read, write bool) error {
	slog.Info(fmt.Sprintf("badgerInteract, [%s]", caller))
	return s.db.Update(func(txn *badger.Txn) error {
		return callback(txn)
	})
}

// GetStruct returns the value for the given key.
func (s *Store) GetStruct(key string, value any) error {
	data, err := s.GetValue(key)
	if err != nil {
		return err
	}
	return Deserialize(value, data)
}

// PutLogEntry put log entry
func (s *Store) PutLogEntry(idxKey uint64, value []byte) error {
	callback := func(txn *badger.Txn) error {
		return txn.Set(logKey(idxKey), value)
	}

	if err := s.badgerInteract("PutLogEntry", callback, false, true); err != nil {
		return err
	}

	s.logIdx = idxKey
	return nil
}

// GetLogEntry get log entry
func (s *Store) GetLogEntry(idxKey uint64) ([]byte, error) {
	var result []byte
	callback := func(txn *badger.Txn) error {
		item, _ := txn.Get(logKey(idxKey))
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		result = val
		return nil
	}

	if err := s.badgerInteract("GetLogEntry", callback, true, false); err != nil {
		return nil, err
	}
	return result, nil
}

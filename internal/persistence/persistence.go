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
		List() ([][]byte, error)
	}

	// BadgerPersistence is a wrapper around BadgerDB.
	BadgerPersistence struct {
		ctx     context.Context
		db      *badger.DB
		keyList map[string]KeyValue
		mutex   *sync.Mutex
	}

	KeyValue struct {
		ID       uint
		Key      uuid.UUID
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
		mutex:   &sync.Mutex{},
		ctx:     ctx,
	}

	wg := &sync.WaitGroup{}

	go b.keyAppender(wg)
	go b.monitorKeys(wg)

	return b, nil
}

func (b *BadgerPersistence) monitorKeys(wg *sync.WaitGroup) {
	wg.Add(1)
	ticker := time.NewTicker(1 * time.Second)

	defer func() {
		ticker.Stop()
		wg.Done()
	}()

	for {
		//// if map is empty and db is not empty, load all keyList from db
		//if len(b.keyList) == 0 {
		//	keyList, err := b.List()
		//	if err != nil {
		//		return nil, err
		//	}
		//
		//	for _, key := range keyList {
		//		b.keyList[string(key)] = KeyValue{
		//			Key:      uuid.New(),
		//			CreateAt: time.Now().UnixNano(),
		//		}
		//	}
		//}

		for {
			select {
			case <-ticker.C:
				log.Info("ticker")
			case <-b.ctx.Done():
				return
			}
		}
	}
}

func (b *BadgerPersistence) keyAppender(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			return
		default:
			keyList, err := b.List()
			if err != nil {
				log.Error(err.Error())
				continue
			}

			for _, key := range keyList {
				b.keyList[string(key)] = KeyValue{
					Key:      uuid.New(),
					CreateAt: time.Now().UnixNano(),
				}
			}
		}
		//data, err := b.serialize(value)
		//if err != nil {
		//	return err
		//}
		//
		//if err = b.db.Update(func(txn *badger.Txn) error {
		//	return txn.Set([]byte("keyList"), data)
		//}); err != nil {
		//	return err
		//}
		//
		//return nil
	}
}

func (b *BadgerPersistence) composeKey(uuidKey *string) KeyValue {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if uuidKey != nil {
		return KeyValue{
			Key:      uuid.MustParse(*uuidKey),
			CreateAt: time.Now().UnixNano(),
		}
	}

	return KeyValue{
		Key:      uuid.New(),
		CreateAt: time.Now().UnixNano(),
	}
}

func (b *BadgerPersistence) appendKey(value KeyValue) error {
	data, err := b.serialize(value)
	if err != nil {
		return err
	}

	if err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("keyList"), data)
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
			item, err := txn.Get(k.Key.NodeID())
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
	return b.SetKey(nil, value)
}

func (b *BadgerPersistence) SetKey(uuidKey *string, value []byte) (string, error) {
	genKey := b.composeKey(uuidKey)
	if err := b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(genKey.Key.NodeID(), value)
	}); err != nil {
		return "", err
	}

	name := genKey.Key.String()
	b.keyList[name] = genKey

	return name, nil
}

func (b *BadgerPersistence) Delete(key string) error {
	if k, ok := b.keyList[key]; ok {
		if err := b.db.Update(func(txn *badger.Txn) error {
			return txn.Delete(k.Key.NodeID())
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
			key := item.KeyCopy(nil)
			keys[string(key)] = KeyValue{
				Key:      uuid.New(),
				CreateAt: time.Now().UnixNano(),
			}
		}

		return nil
	})

	return keys, err
}

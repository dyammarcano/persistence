//go:build windows

package persistence

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
	"time"
)

var (
	persistence *CachePersistence
	value       = []byte("hello world from go guys!")
)

func init() {
	path := filepath.Clean("../../test.db")
	var err error
	persistence, err = NewBadgerPersistence(context.TODO(), path)
	if err != nil {
		panic(err)
	}
}

func TestBadgerPersistence_SetStructAsync(t *testing.T) {
	msgObj := &struct {
		Message string
		Hash    string
	}{
		Message: "hello world",
		Hash:    uuid.NewString(),
	}

	callbackFn := func(key string, err error) {
		assert.NoErrorf(t, err, "expected error, but got '%s'", err)

		got := &struct {
			Message string
			Hash    string
		}{}

		err = persistence.GetStruct(key, got)
		assert.NoErrorf(t, err, "expected error, but got '%s'", err)

		assert.Equalf(t, msgObj.Hash, got.Hash, "expected '%s', but got '%s'", msgObj.Hash, got.Hash)
		fmt.Printf("key: %s, value: %v\n", key, got)
	}

	start := time.Now()
	persistence.SetStructAsync(msgObj, callbackFn)
	fmt.Printf("SetStructAsync time: %s\n", time.Since(start))

	<-time.After(5 * time.Second)
}

func TestBadgerPersistence_Get(t *testing.T) {
	key, err := persistence.SetValue(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	got, err := persistence.GetValue(key)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.Equalf(t, value, got, "expected '%s', but got '%s'", string(value), string(got))

	fmt.Printf("key: %s, value: %s\n", key, got)
}

func TestBadgerPersistence_ListKeys(t *testing.T) {
	keys := persistence.ListKeys()
	assert.NotNilf(t, keys, "expected keys, but got '%s'", keys)
}

func TestBadgerPersistence_Set1(t *testing.T) {
	key, err := persistence.SetValue(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.NotNilf(t, key, "expected key, but got '%s'", key)
}

func TestBadgerPersistence_Set1_000(t *testing.T) {
	for i := 0; i < 1000; i++ {
		key, err := persistence.SetValue(value)
		assert.NoErrorf(t, err, "expected error, but got '%s'", err)
		assert.NotNilf(t, key, "expected key, but got '%s'", key)
	}

	if persistence.Len() < 1000 {
		t.Errorf("expected 1000 keys, but got '%d'", persistence.Len())
	}

	<-time.After(5 * time.Second)
}

func TestBadgerPersistence_Set(t *testing.T) {
	key, err := persistence.SetValue(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.NotNilf(t, key, "expected key, but got '%s'", key)
}

func TestBadgerPersistence_Delete(t *testing.T) {
	key, err := persistence.SetValue(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	err = persistence.Delete(key)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	_, err = persistence.GetValue(key)
	assert.Equalf(t, ErrKeyNotFound, err, "expected '%s', but got '%s'", ErrKeyNotFound, err)
}

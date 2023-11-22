//go:build windows

package persistence

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

var (
	persistence *BadgerPersistence
	value       = []byte("hello world from go guys!")
)

func TestMain(m *testing.M) {
	var err error
	persistence, err = NewBadgerPersistence(context.TODO(), filepath.Clean("../../test.db"))
	if err != nil {
		panic(err)
	}

	code := m.Run()
	persistence.Close()
	err = os.RemoveAll(filepath.Clean("../../test.db"))
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestBadgerPersistence_Get(t *testing.T) {
	response, err := persistence.Set(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	got, err := persistence.Get(response.String)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.Equalf(t, value, got.Payload, "expected '%s', but got '%s'", string(value), string(got.Payload))

	fmt.Printf("key: %s, value: %s\n", response.String, got.Payload)
}

func TestBadgerPersistence_ListKeys(t *testing.T) {
	keys, err := persistence.ListKeys()
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.NotNilf(t, keys, "expected keys, but got '%s'", keys)
}

func TestBadgerPersistence_Set1(t *testing.T) {
	key, err := persistence.Set(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.NotNilf(t, key, "expected key, but got '%s'", key.String)
}

func TestBadgerPersistence_Set1_000(t *testing.T) {
	for i := 0; i < 1000; i++ {
		key, err := persistence.Set(value)
		assert.NoErrorf(t, err, "expected error, but got '%s'", err)
		assert.NotNilf(t, key, "expected key, but got '%s'", key.String)
	}

	if persistence.Len() < 1000 {
		t.Errorf("expected 1000 keys, but got '%d'", persistence.Len())
	}
}

func TestBadgerPersistence_Set(t *testing.T) {
	key, err := persistence.Set(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.NotNilf(t, key, "expected key, but got '%s'", key.String)
}

func TestBadgerPersistence_Delete(t *testing.T) {
	key, err := persistence.Set(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	err = persistence.Delete(key.String)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	_, err = persistence.Get(key.String)
	assert.Equalf(t, ErrKeyNotFound, err, "expected '%s', but got '%s'", ErrKeyNotFound, err)
}

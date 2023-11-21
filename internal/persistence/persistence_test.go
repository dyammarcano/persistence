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

	os.Exit(code)
}

func TestBadgerPersistence_Get(t *testing.T) {
	response, err := persistence.Set(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	got, err := persistence.Get(response.Key)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.Equalf(t, value, got.Payload, "expected '%s', but got '%s'", string(value), string(got.Payload))

	fmt.Printf("key: %s, value: %s\n", response.Key, got.Payload)
}

//
//func TestBadgerPersistence_ListKeys(t *testing.T) {
//	keys, err := persistence.loadKeys()
//	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
//
//	fmt.Println(keys)
//}
//
//func TestBadgerPersistence_Set1(t *testing.T) {
//	key, err := persistence.Set(value)
//	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
//	assert.NotNilf(t, key, "expected key, but got '%s'", key)
//}
//
//func TestBadgerPersistence_Set100(t *testing.T) {
//	for i := 0; i < 100; i++ {
//		key, err := persistence.Set(value)
//		assert.NoErrorf(t, err, "expected error, but got '%s'", err)
//		assert.NotNilf(t, key, "expected key, but got '%s'", key)
//	}
//
//	assert.Equalf(t, 100, len(persistence.keyList), "expected 100 keys, but got '%d'", len(persistence.keyList))
//}
//
//func TestBadgerPersistence_Set(t *testing.T) {
//	key, err := persistence.Set(value)
//	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
//	assert.NotNilf(t, key, "expected key, but got '%s'", key)
//}
//
//func TestBadgerPersistence_Delete(t *testing.T) {
//	key, err := persistence.Set(value)
//	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
//
//	err = persistence.Delete(key)
//	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
//
//	_, err = persistence.Get(key)
//	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
//}

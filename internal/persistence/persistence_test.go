package persistence

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var persistence *BadgerPersistence

func TestMain(m *testing.M) {
	var err error
	persistence, err = NewBadgerPersistence(context.TODO(), "./test.db")
	if err != nil {
		panic(err)
	}

	code := m.Run()

	persistence.Close()
	_ = os.RemoveAll("./test.db")

	os.Exit(code)
}

func TestBadgerPersistence_Get(t *testing.T) {
	value := []byte("testValue")

	key, err := persistence.Set(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	got, err := persistence.Get(key)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.Equalf(t, value, got, "expected '%s', but got '%s'", string(value), string(got))

	fmt.Printf("key: %s, value: %s\n", key, got)
}

func TestBadgerPersistence_Set(t *testing.T) {
	value := []byte("testValue")

	key, err := persistence.Set(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
	assert.NotNilf(t, key, "expected key, but got '%s'", key)
}

func TestBadgerPersistence_Delete(t *testing.T) {
	value := []byte("testValue")

	key, err := persistence.Set(value)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	err = persistence.Delete(key)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	_, err = persistence.Get(key)
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
}

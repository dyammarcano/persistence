package persistence

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	persistence *Store
	value       = []byte("hello world from go guys!")
)

func init() {
	var err error
	persistence, err = NewBadgerPersistence(context.Background(), true, "")
	if err != nil {
		panic(err)
	}
	<-time.After(10 * time.Millisecond) // wait for the keys to be loaded
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

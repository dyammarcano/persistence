package persistence

import (
	"github.com/google/uuid"
	"os"
	"testing"
)

var persistence *BadgerPersistence

func generateId() []byte {
	return uuid.New().NodeID()
}

func TestMain(m *testing.M) {
	var err error
	persistence, err = NewBadgerPersistence("./test.db")
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = persistence.Close()
	_ = os.RemoveAll("./test.db")

	os.Exit(code)
}

func TestBadgerPersistence_Get(t *testing.T) {
	key := generateId()
	value := []byte("testValue")

	if err := persistence.Set(key, value); err != nil {
		t.Fatal(err)
	}

	got, err := persistence.Get(key)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(value) {
		t.Fatalf("expected '%s', but got '%s'", string(value), string(got))
	}
}

func TestBadgerPersistence_Set(t *testing.T) {
	key := generateId()
	value := []byte("testValue")

	if err := persistence.Set(key, value); err != nil {
		t.Fatal(err)
	}
}

func TestBadgerPersistence_Update(t *testing.T) {
	key := generateId()
	value := []byte("testValue")
	newValue := []byte("newValue")

	if err := persistence.Set(key, value); err != nil {
		t.Fatal(err)
	}

	if err := persistence.Update(key, newValue); err != nil {
		t.Fatal(err)
	}

	got, err := persistence.Get(key)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(newValue) {
		t.Fatalf("expected '%s', but got '%s'", string(newValue), string(got))
	}
}

func TestBadgerPersistence_Delete(t *testing.T) {
	key := generateId()
	value := []byte("testValue")

	if err := persistence.Set(key, value); err != nil {
		t.Fatal(err)
	}

	if err := persistence.Delete(key); err != nil {
		t.Fatal(err)
	}

	if _, err := persistence.Get(key); err == nil {
		t.Fatalf("expected error, but got nil")
	}
}

func TestBadgerPersistence_List(t *testing.T) {
	exampleKeys := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
		[]byte("key3"),
	}

	for _, key := range exampleKeys {
		err := persistence.Set(key, []byte("value"))
		if err != nil {
			t.Fatal(err)
		}
	}

	keys, err := persistence.List()
	if err != nil {
		t.Fatal(err)
	}

	for _, key := range exampleKeys {
		found := false
		for _, listedKey := range keys {
			if string(key) == string(listedKey) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("key %s not found in the list", string(key))
		}
	}
}

func TestBadgerPersistence_Close(t *testing.T) {
	if err := persistence.Close(); err != nil {
		t.Fatal(err)
	}
}

//go:build windows

package persistence

import (
	"context"
	v1 "documentDatabaseTest/internal/models/v1"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	persistence *CachePersistence
	value       = []byte("hello world from go guys!")
)

func measureTime(fn func()) {
	start := time.Now()
	fn()
	fmt.Printf("time: %s\n", time.Since(start))
}

func TestMain(m *testing.M) {
	path := filepath.Clean("../../test.db")
	var err error
	persistence, err = NewBadgerPersistence(context.TODO(), path)
	if err != nil {
		panic(err)
	}

	code := m.Run()
	persistence.Close()
	//err = os.RemoveAll(path)
	//if err != nil {
	//	panic(err)
	//}
	os.Exit(code)
}

func TestBadgerPersistence_GetStruct(t *testing.T) {
	msgObj := &v1.OperationStatus{
		RuntimeVersion: "1.0.0",
		CreateAt:       1700626863623721700,
		UpdateAt:       1700626863623721700,
		ID:             "831e1650-001e-001b-66ab-eeb76e069631",
		OperationID:    "e8e564fd-38f5-4684-9581-c30f6c25213a",
		Status:         "Failed",
		CorrelationID:  "831e1650-001e-001b-66ab-eeb76e000000",
		FileInfo: &v1.FileInfo{
			ETag:          "0x8D4BCC2E4835CD0",
			ContentType:   "application/octet-stream",
			ContentLength: 524288,
			Hash:          "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		Stages: &v1.Stages{
			Stage0: &v1.Stage{
				StartTime: "2017-06-26T18:41:00.9584103Z",
				EndTime:   "2017-06-26T18:41:00.9584103Z",
				Message:   "The request is invalid.",
				InnerError: &v1.InnerError{
					Date:    "2017-06-26T18:41:00",
					Code:    "InvalidRequest",
					Message: "File not meet the requirements.",
				},
			},
			Stage1: &v1.Stage{
				StartTime: "2017-06-26T18:41:00.9584103Z",
				EndTime:   "2017-06-26T18:41:00.9584103Z",
				Message:   "The request is invalid.",
				InnerError: &v1.InnerError{
					Date:    "2017-06-26T18:41:00",
					Code:    "InvalidRequest",
					Message: "File not meet the requirements.",
				},
			},
		},
	}

	start := time.Now()
	key, err := persistence.SetStruct(msgObj)
	fmt.Printf("SetStruct time: %s\n", time.Since(start))
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	got := &v1.OperationStatus{}
	start2 := time.Now()
	err = persistence.GetStruct(key, got)
	fmt.Printf("GetStruct time: %s\n", time.Since(start2))
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)

	assert.Equalf(t, msgObj.FileInfo.ETag, got.FileInfo.ETag, "expected '%s', but got '%s'", msgObj.FileInfo.ETag, got.FileInfo.ETag)
	fmt.Printf("key: %s, value: %v\n", key, got)
}

func TestBadgerPersistence_SetStructAsync(t *testing.T) {
	msgObj := &v1.OperationStatus{
		RuntimeVersion: "1.0.0",
		CreateAt:       1700626863623721700,
		UpdateAt:       1700626863623721700,
		ID:             "831e1650-001e-001b-66ab-eeb76e069631",
		OperationID:    "e8e564fd-38f5-4684-9581-c30f6c25213a",
		Status:         "Failed",
		CorrelationID:  "831e1650-001e-001b-66ab-eeb76e000000",
		FileInfo: &v1.FileInfo{
			ETag:          "0x8D4BCC2E4835CD0",
			ContentType:   "application/octet-stream",
			ContentLength: 524288,
			Hash:          "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		Stages: &v1.Stages{
			Stage0: &v1.Stage{
				StartTime: "2017-06-26T18:41:00.9584103Z",
				EndTime:   "2017-06-26T18:41:00.9584103Z",
				Message:   "The request is invalid.",
				InnerError: &v1.InnerError{
					Date:    "2017-06-26T18:41:00",
					Code:    "InvalidRequest",
					Message: "File not meet the requirements.",
				},
			},
			Stage1: &v1.Stage{
				StartTime: "2017-06-26T18:41:00.9584103Z",
				EndTime:   "2017-06-26T18:41:00.9584103Z",
				Message:   "The request is invalid.",
				InnerError: &v1.InnerError{
					Date:    "2017-06-26T18:41:00",
					Code:    "InvalidRequest",
					Message: "File not meet the requirements.",
				},
			},
		},
	}

	callbackFn := func(key string, err error) {
		assert.NoErrorf(t, err, "expected error, but got '%s'", err)

		got := &v1.OperationStatus{}

		err = persistence.GetStruct(key, got)
		assert.NoErrorf(t, err, "expected error, but got '%s'", err)

		assert.Equalf(t, msgObj.FileInfo.ETag, got.FileInfo.ETag, "expected '%s', but got '%s'", msgObj.FileInfo.ETag, got.FileInfo.ETag)
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
	assert.Equalf(t, value, got.Payload, "expected '%s', but got '%s'", string(value), string(got.Payload))

	fmt.Printf("key: %s, value: %s\n", key, got.Payload)
}

func TestBadgerPersistence_ListKeys(t *testing.T) {
	keys, err := persistence.ListKeys()
	assert.NoErrorf(t, err, "expected error, but got '%s'", err)
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

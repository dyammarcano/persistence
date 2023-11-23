package documentDatabaseTest

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"github.com/dyammarcano/base58"
	"strings"
)

const (
	BDGDATPREFIX = "dat:"
	BDGLOGPREFIX = "log:"
)

var (
	dbDatPrefix = []byte(BDGDATPREFIX)
	dbLogPrefix = []byte(BDGLOGPREFIX)
)

// generateRandomKey generate random key with fix length of 10 bits + prefix.
func generateRandomKey() []byte {
	var vk = make([]byte, 10)
	if _, err := rand.Read(vk); err != nil {
		return nil
	}
	return dataKey(vk)
}

// composeKey generate a new key.
func composeKey(key []byte) *Key {
	return &Key{
		Key:    key,
		String: EncodeKey(key),
	}
}

// EncodeKey encodes a key and return a string with 27 characters.
func EncodeKey(key []byte) string {
	return strings.ToUpper(base58.StdEncoding.EncodeToString(sha1.New().Sum(key)))[0:27]
}

// Serialize serializes a value.
func Serialize(value any) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Deserialize deserializes a value.
func Deserialize(obj any, val []byte) error {
	if err := gob.NewDecoder(bytes.NewReader(val)).Decode(obj); err != nil {
		return err
	}
	return nil
}

// logKey returns the key used to store the log entry for the given index.
func logKey(idxKey uint64) []byte {
	v := fmt.Sprintf("%s%d", dbLogPrefix, idxKey)
	return []byte(v)
}

// dataKey returns the key used to store the data for the given key.
func dataKey(rawKey []byte) []byte {
	v := fmt.Sprintf("%s%s", dbDatPrefix, rawKey)
	return []byte(v)
}

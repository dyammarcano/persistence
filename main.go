package main

import (
	"context"
	"documentDatabaseTest/internal/persistence"
	"encoding/json"
	"github.com/caarlos0/log"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"path/filepath"
	"time"
)

func mockData() ([]byte, error) {
	msgObj := &struct {
		Message string
		Hash    string
	}{
		Message: "hello world",
		Hash:    uuid.NewString(),
	}

	log.Infof("Mock data: %s", msgObj.Hash)

	data, err := json.Marshal(msgObj)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func main() {
	ctx := context.TODO()
	persistent, err := persistence.NewBadgerPersistence(ctx, filepath.Clean("./web.db"))
	if err != nil {
		panic(err)
	}
	defer persistent.Close()

	router := mux.NewRouter()

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		log.Infof("Start inject random data")

		for {
			select {
			case <-ticker.C:
				data, err := mockData()
				if err != nil {
					log.Fatal(err.Error())
				}
				if _, err := persistent.SetValue(data); err != nil {
					log.Fatal(err.Error())
				}
			}
		}
	}()

	log.Infof("Start web server")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err.Error())
	}
}

func generateId() []byte {
	return uuid.New().NodeID()
}

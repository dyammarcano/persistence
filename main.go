package main

import (
	"context"
	"crypto/sha256"
	v1 "documentDatabaseTest/internal/models/v1"
	"documentDatabaseTest/internal/persistence"
	"github.com/caarlos0/log"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"path/filepath"
	"time"
)

func mockData() ([]byte, error) {
	msgObj := &v1.OperationStatus{
		RuntimeVersion: "1.0.0",
		ID:             uuid.NewString(),
		OperationID:    uuid.NewString(),
		Status:         "Failed",
		CorrelationID:  uuid.NewString(),
		FileInfo: &v1.FileInfo{
			ETag:          "0x8D4BCC2E4835CD0",
			ContentType:   "application/octet-stream",
			ContentLength: 524288,
			Hash:          string(sha256.New().Sum(uuid.NodeID())),
		},
		Stages: &v1.Stages{
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
			Stage2: &v1.Stage{
				StartTime: "2017-06-26T18:41:00.9584103Z",
				EndTime:   "2017-06-26T18:41:00.9584103Z",
				Message:   "The request is invalid.",
				InnerError: &v1.InnerError{
					Date:    "2017-06-26T18:41:00",
					Code:    "InvalidRequest",
					Message: "File not meet the requirements.",
				},
			},
			Stage3: &v1.Stage{
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

	log.Infof("Mock data: %s", msgObj.OperationID)

	return msgObj.SerializeBytes()
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

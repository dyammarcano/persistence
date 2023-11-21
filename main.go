package main

import (
	"documentDatabaseTest/internal/persistence"
	"github.com/caarlos0/log"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

var instance *persistence.BadgerPersistence

func init() {
	var err error
	instance, err = persistence.NewBadgerPersistence("./test.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer instance.Close()
}

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/", HomeHandler)

	log.Info("Starting server on port 8080")
	if err := http.ListenAndServe(":8080", route); err != nil {
		log.Fatal(err.Error())
	}
}

func injectRandomData() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var paragraphs []string
	lap := 0

	for {
		lap++
		select {
		case <-ticker.C:
			key := generateId()
			value := []byte(paragraphs[len(paragraphs)%lap])
			if err := instance.Set(key, value); err != nil {
				log.Fatal(err.Error())
			}
		}
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func generateId() []byte {
	return uuid.New().NodeID()
}

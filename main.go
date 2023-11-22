package main

import (
	"context"
	"documentDatabaseTest/internal/persistence"
	"github.com/caarlos0/log"
	"github.com/google/uuid"
	"path/filepath"
	"time"
)

//var instance *persistence.BadgerPersistence
//
//func init() {
//	var err error
//	instance, err = persistence.NewBadgerPersistence("./test.db")
//	if err != nil {
//		log.Fatal(err.Error())
//	}
//	defer instance.Close()
//}

func main() {
	ctx := context.TODO()
	persistent, err := persistence.NewBadgerPersistence(ctx, filepath.Clean("./web.db"))
	if err != nil {
		panic(err)
	}
	defer persistent.Close()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		log.Infof("Start inject random data")

		for {
			select {
			case <-ticker.C:
				if _, err := persistent.Set(generateId()); err != nil {
					log.Fatal(err.Error())
				}
			}
		}
	}()

	persistent.StartWebInterface(":8080")
}

//		route := mux.NewRouter()
//		route.HandleFunc("/", HomeHandler)
//
//		log.Info("Starting server on port 8080")
//		if err := http.ListenAndServe(":8080", route); err != nil {
//			log.Fatal(err.Error())
//		}
//	}
//
//	func injectRandomData() {
//		ticker := time.NewTicker(1 * time.Second)
//		defer ticker.Stop()
//
//		var paragraphs []string
//		lap := 0
//
//		for {
//			lap++
//			select {
//			case <-ticker.C:
//				key := generateId()
//				value := []byte(paragraphs[len(paragraphs)%lap])
//				if err := instance.Set(key, value); err != nil {
//					log.Fatal(err.Error())
//				}
//			}
//		}
//	}
//
//	func HomeHandler(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusOK)
//	}

func generateId() []byte {
	return uuid.New().NodeID()
}

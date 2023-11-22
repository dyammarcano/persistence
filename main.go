package main

import (
	"context"
	"documentDatabaseTest/internal/persistence"
	"github.com/caarlos0/log"
	"github.com/google/uuid"
	"path/filepath"
	"time"
)

func main() {
	ctx := context.TODO()
	persistent, err := persistence.NewBadgerPersistence(ctx, filepath.Clean("./web.db"))
	if err != nil {
		panic(err)
	}
	defer persistent.Close()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
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

	persistent.RegisterPersistenceWebInterface(":8080")
}

func generateId() []byte {
	return uuid.New().NodeID()
}

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/evt/callback/internal/services/testerservice"

	"github.com/evt/callback/internal/model"
	"github.com/evt/callback/internal/pg"
	"github.com/evt/callback/internal/repositories/objectrepo"
	"github.com/evt/callback/internal/services/callbackservice"

	"github.com/evt/callback/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// default context
	defaultCtx := context.Background()

	// config
	cfg := config.Get()

	// connect to Postgres
	pgDB, err := pg.Dial()
	if err != nil {
		return fmt.Errorf("pgdb.Dial failed: %w", err)
	}

	// run Postgres migrations
	if pgDB != nil {
		log.Println("Running PostgreSQL migrations")
		if err := runPgMigrations(); err != nil {
			return fmt.Errorf("runPgMigrations failed: %w", err)
		}
	}

	// object repository
	objectRepo := objectrepo.New(pgDB)

	// delete objects in the database when they have not been received for more than 30 seconds
	go func() {
		if err := objectRepo.CleanExpiredObjects(defaultCtx); err != nil {
			log.Fatal(err)
		}
	}()

	// callback service
	callbackService := callbackservice.New(objectRepo)

	// tester service to get object details (allow waiting for 30 seconds max till all requests completed)
	testerService := testerservice.New(time.Second * 60)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		var request model.CallbackRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %s", err), http.StatusBadRequest)
		}

		if len(request.ObjectIDs) == 0 {
			http.Error(w, "no object IDs provided", http.StatusBadRequest)
		}

		// this is a group of requests back to tester service for object details
		var wg sync.WaitGroup
		receivedObjects := make(chan model.TesterObject, len(request.ObjectIDs))

		for i := range request.ObjectIDs {
			wg.Add(1)

			objectID := request.ObjectIDs[i]

			//log.Printf("=> Next object ID: %d\n", objectID)

			go func() {
				defer wg.Done()

				object, err := testerService.GetObject(objectID)
				if err != nil {
					log.Printf("[id: %d, total: %d] testerService.GetObject failed: %s\n", object.ID, len(request.ObjectIDs), err)

					return
				}

				//log.Printf("[id: %d, total: %d] testerService.GetObject passed (online=%t)\n", object.ID, len(request.ObjectIDs), object.Online)

				receivedObjects <- object

			}()
		}

		go func() {
			wg.Wait()
			close(receivedObjects)
		}()

		var totalUpdated, totalReceived int

		for object := range receivedObjects {
			totalReceived++

			if !object.Online {
				continue
			}

			callbackService.UpdateObject(defaultCtx, &model.Object{
				ID: object.ID,
			})

			totalUpdated++
		}

		log.Printf("objects: received=%d, updated(=online) = %d\n", totalReceived, totalUpdated)

		w.Write([]byte("ok"))
	})

	log.Printf("Running HTTP server on %s\n", cfg.HTTPAddr)

	go func() { _ = http.ListenAndServe(cfg.HTTPAddr, nil) }()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("closing")

	return nil
}

// runPgMigrations runs Postgres migrations
func runPgMigrations() error {
	cfg := config.Get()

	if cfg.PgMigrationsPath == "" {
		return nil
	}

	if cfg.PgURL == "" {
		return errors.New("No cfg.PgURL provided")
	}

	m, err := migrate.New(
		cfg.PgMigrationsPath,
		cfg.PgURL,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

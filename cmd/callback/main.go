package main

import (
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
	// config
	cfg := config.Get()

	// connect to Postgres
	pgDB, err := pg.Dial()
	if err != nil {
		return fmt.Errorf("pgdb.Dial failed: %w", err)
	}

	// Run Postgres migrations
	if pgDB != nil {
		log.Println("Running PostgreSQL migrations")
		if err := runPgMigrations(); err != nil {
			return fmt.Errorf("runPgMigrations failed: %w", err)
		}
	}

	// object repository
	objectRepo := objectrepo.New(pgDB)

	// callback service
	callbackService := callbackservice.New(objectRepo)

	// tester service (to get object)
	testerService := testerservice.New(time.Second * 3)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var request model.CallbackRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, fmt.Sprintf("invalid request: %s", err), http.StatusBadRequest)
		}

		if len(request.ObjectIDs) == 0 {
			http.Error(w, "no object IDs provided", http.StatusBadRequest)
		}

		var wg sync.WaitGroup
		receivedObjects := make(chan model.TesterObject, len(request.ObjectIDs))

		for i := range request.ObjectIDs {
			wg.Add(1)

			go func(objectID uint) {
				defer wg.Done()

				object, err := testerService.GetObject()
				if err != nil {
					log.Printf("[%d of %d] testerService.GetObject failed: %s\n", objectID, len(request.ObjectIDs), err)
					return
				}

				log.Printf("[%d of %d] testerService.GetObject passed\n", objectID, len(request.ObjectIDs))

				receivedObjects <- *object

			}(request.ObjectIDs[i])
		}

		go func() {
			wg.Wait()
			close(receivedObjects)
		}()

		for _ = range receivedObjects {

		}

		for i := range request.ObjectIDs {
			callbackService.CreateObject(r.Context(), &model.Object{
				ID: request.ObjectIDs[i],
			})
		}

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

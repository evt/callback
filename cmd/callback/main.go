package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evt/callback/internal/handlers/callbackhandler"

	"github.com/evt/callback/internal/services/testerservice"

	"github.com/evt/callback/internal/pg"
	"github.com/evt/callback/internal/repositories/objectrepo"
	"github.com/evt/callback/internal/services/objectservice"

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

	// clean architecture: handler -> service -> repository

	// repository init
	objectRepo := objectrepo.New(pgDB)

	// delete objects in the database when they have not been received for more than 30 seconds
	go func() {
		if err := objectRepo.CleanExpiredObjects(defaultCtx); err != nil {
			log.Fatal(err)
		}
	}()

	// service init
	objectService := objectservice.New(objectRepo)
	testerService := testerservice.New(time.Second * 60)

	// handler init
	callbackHandler := callbackhandler.New(objectService, testerService)

	// routes
	http.HandleFunc("/callback", callbackHandler.Post)

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

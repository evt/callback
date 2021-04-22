package pg

import (
	"errors"
	"fmt"
	"time"

	"github.com/evt/callback/config"

	"github.com/go-pg/pg/v10"
)

// Timeout is a Postgres timeout
const Timeout = 5

// DB is a shortcut structure to a Postgres DB
type DB struct {
	*pg.DB
}

// Dial creates new database connection to postgres
func Dial() (*DB, error) {
	cfg := config.Get()
	if cfg.PgURL == "" {
		return nil, errors.New("no postgres URL provided")
	}
	pgOpts, err := pg.ParseURL(cfg.PgURL)
	if err != nil {
		return nil, err
	}

	pgDB := pg.Connect(pgOpts)

	_, err = pgDB.Exec("SELECT 1")
	if err != nil {
		return nil, fmt.Errorf("pgDB.Exec failed: %w", err)
	}

	pgDB.WithTimeout(time.Second * time.Duration(Timeout))

	return &DB{pgDB}, nil
}

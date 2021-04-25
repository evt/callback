package pg

import (
	"errors"
	"fmt"
	"log"
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

	// run test select query to make sure PostgreSQL is up and running
	var attempt uint

	const maxAttempts = 10

	for {
		attempt++

		log.Printf("[PostgreSQL.Dial] (Ping attempt %d) SELECT 1\n", attempt)

		_, err = pgDB.Exec("SELECT 1")
		if err != nil {
			log.Printf("[PostgreSQL.Dial] (Ping attempt %d) error: %s\n", attempt, err)

			if attempt < maxAttempts {
				time.Sleep(1 * time.Second)

				continue
			}

			return nil, fmt.Errorf("pgDB.Exec failed: %w", err)
		}

		log.Printf("[PostgreSQL.Dial] (Ping attempt %d) OK\n", attempt)

		break
	}

	pgDB.WithTimeout(time.Second * time.Duration(Timeout))

	return &DB{pgDB}, nil
}

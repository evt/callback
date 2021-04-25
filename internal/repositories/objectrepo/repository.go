package objectrepo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/evt/callback/internal/model"

	"github.com/evt/callback/internal/pg"
)

// ObjectRepository is a object repository.
type ObjectRepository struct {
	db *pg.DB
}

// New creates a new vessel repository.
func New(db *pg.DB) *ObjectRepository {
	return &ObjectRepository{
		db: db,
	}
}

// UpdateObject creates new object in Dynamo DB.
func (repo *ObjectRepository) UpdateObject(ctx context.Context, object *model.DBObject) error {
	_, err := repo.db.
		WithContext(ctx).
		Model(object).
		OnConflict("(id) DO UPDATE").
		Set("last_seen = now()").
		Insert()
	if err != nil {
		return fmt.Errorf("SQL insert failed: %w", err)
	}

	return nil
}

// CleanExpiredObjects removes objects with last_seen > 30 sec every 30 sec
func (repo *ObjectRepository) CleanExpiredObjects(ctx context.Context) error {
	const (
		ttl        = 30
		tickPeriod = 30
	)

	ticker := time.NewTicker(time.Second * tickPeriod)
	defer ticker.Stop()

	for {
		<-ticker.C

		result, err := repo.db.
			WithContext(ctx).
			Model((*model.DBObject)(nil)).
			Where(fmt.Sprintf("last_seen < (now() - '%d seconds'::interval)", ttl)).
			Delete()
		if err != nil {
			return fmt.Errorf("SQL delete failed: %w", err)
		}

		if result.RowsAffected() > 0 {
			log.Printf("[CleanExpiredObjects] Deleted %d objects last seen over %d sec ago\n", result.RowsAffected(), ttl)
		}
	}

	return nil
}

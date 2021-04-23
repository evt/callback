package objectrepo

import (
	"context"

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
func (repo *ObjectRepository) UpdateObject(ctx context.Context, object *model.Object) error {
	_, err := repo.db.
		WithContext(ctx).
		Model(object).
		OnConflict("(id) DO UPDATE").
		Set("last_seen = now()").
		Insert()
	if err != nil {
		return err
	}

	return nil
}

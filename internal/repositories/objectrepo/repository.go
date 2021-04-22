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

// CreateObject creates new object in Dynamo DB.
func (repo *ObjectRepository) CreateObject(ctx context.Context, object *model.Object) error {
	_, err := repo.db.
		WithContext(ctx).
		Model(object).
		OnConflict("DO NOTHING").
		Insert()
	if err != nil {
		return err
	}

	return nil
}

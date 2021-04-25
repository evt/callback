package objectservice

import (
	"context"

	"github.com/evt/callback/internal/model"
)

// ObjectRepository is an object repository
type ObjectRepository interface {
	UpdateObject(context.Context, *model.DBObject) error
	// GetObject(context.Context, string) (*model.DBObject, error)
}

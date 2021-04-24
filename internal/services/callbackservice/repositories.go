package callbackservice

import (
	"context"

	"github.com/evt/callback/internal/model"
)

// ObjectRepository is an object repository
type ObjectRepository interface {
	UpdateObject(context.Context, *model.Object) error
	// GetObject(context.Context, string) (*model.Object, error)
}

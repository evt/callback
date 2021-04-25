//go:generate mockgen -destination=./mocks.go -source=./services.go -package=callbackhandller

package callbackhandler

import (
	"context"

	"github.com/evt/callback/internal/e"

	"github.com/evt/callback/internal/model"
)

// ObjectService is a service to update object details in PostgreSQL.
type ObjectService interface {
	UpdateObject(ctx context.Context, object *model.DBObject) e.Error
}

// TesterService is a service to get object details from tester service.
type TesterService interface {
	GetObject(ctx context.Context, objectID uint) (model.TesterObject, e.Error)
}

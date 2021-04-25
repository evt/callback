package objectservice

import (
	"context"
	"time"

	"github.com/evt/callback/internal/e"

	"github.com/evt/callback/internal/model"
)

// ObjectService is a callback service.
type ObjectService struct {
	objectRepo ObjectRepository
}

// New creates a new callback service.
func New(repo ObjectRepository) *ObjectService {
	return &ObjectService{
		objectRepo: repo,
	}
}

// UpdateObject creates new object
func (svc *ObjectService) UpdateObject(ctx context.Context, object *model.DBObject) e.Error {
	if object == nil {
		return e.NewInternal("no object provided")
	}

	if object.ID == 0 {
		return e.NewBadRequest("no object ID provided")
	}

	object.LastSeen = time.Now().UTC()

	err := svc.objectRepo.UpdateObject(ctx, object)
	if err != nil {
		return e.NewInternalf("failed creating object in repo: %s", err)
	}

	return nil
}

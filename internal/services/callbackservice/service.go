package callbackservice

import (
	"context"
	"time"

	"github.com/evt/callback/internal/e"

	"github.com/evt/callback/internal/model"
)

// CallbackService is a callback service.
type CallbackService struct {
	objectRepo ObjectRepository
}

// New creates a new callback service.
func New(repo ObjectRepository) *CallbackService {
	return &CallbackService{
		objectRepo: repo,
	}
}

// UpdateObject creates new object
func (svc *CallbackService) UpdateObject(ctx context.Context, object *model.Object) e.Error {
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

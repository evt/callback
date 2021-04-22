package callbackservice

import (
	"context"
	"time"

	"github.com/evt/callback/internal/error"

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

// CreateObject creates new object
func (svc *CallbackService) CreateObject(ctx context.Context, object *model.Object) error.Error {
	if object == nil {
		return error.NewInternal("no object provided")
	}

	if object.ID == 0 {
		return error.NewBadRequest("no object ID provided")
	}

	object.LastSeen = time.Now().UTC()

	err := svc.objectRepo.CreateObject(ctx, object)
	if err != nil {
		return error.NewInternalf("failed creating object in repo: %s", err)
	}

	return nil
}

package testerservice

import (
	"time"

	"github.com/evt/callback/internal/error"
	"github.com/evt/callback/internal/model"
)

// ObjectService is a object service.
type ObjectService struct {
	timeout time.Duration
	limiter chan struct{}
}

// New creates a new object service.
func New(timeout time.Duration) *ObjectService {
	return &ObjectService{
		timeout: timeout,
		limiter: make(chan struct{}, 2), // max 100 parallel requests to tester
	}
}

// GetObject fetches object from tester service
func (svc *ObjectService) GetObject() (*model.TesterObject, error.Error) {
	select {
	case <-time.After(svc.timeout):
		return nil, error.NewInternal("get object timeout")
	case svc.limiter <- struct{}{}:
	}
	defer func() {
		<-svc.limiter
	}()

	time.Sleep(5 * time.Second)

	return &model.TesterObject{}, nil
}

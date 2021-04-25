package testerservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/evt/callback/internal/e"

	"github.com/evt/callback/internal/model"
)

// tester object details endpoint response time is between 300ms and 4s, so 5 secs shouldl be enough
var httpClient = &http.Client{Timeout: 5 * time.Second}

// ObjectService is a object service.
type ObjectService struct {
	timeout time.Duration
	limiter chan struct{}
}

// New creates a new object service.
func New(timeout time.Duration) *ObjectService {
	return &ObjectService{
		timeout: timeout,                  // timeout for GET http://tester:9010/objects/<object ID>
		limiter: make(chan struct{}, 200), // max 200 parallel requests to tester service
	}
}

// GetObject fetches object from tester service
func (svc *ObjectService) GetObject(ctx context.Context, objectID uint) (model.TesterObject, e.Error) {
	if objectID == 0 {
		return model.TesterObject{}, e.NewBadRequest("no object ID provided")
	}

	select {
	case <-time.After(svc.timeout):
		return model.TesterObject{}, e.NewInternalf("get object timeout (%.0f secs)", svc.timeout.Seconds())
	case svc.limiter <- struct{}{}:
	}
	defer func() {
		<-svc.limiter
	}()

	url := fmt.Sprintf("http://tester:9010/objects/%d", objectID)

	response, err := httpClient.Get(url)
	if err != nil {
		return model.TesterObject{}, e.NewInternalf("get object details failed: %s", err)
	}
	defer response.Body.Close()

	var testerObject model.TesterObject
	if err := json.NewDecoder(response.Body).Decode(&testerObject); err != nil {
		return model.TesterObject{}, e.NewInternalf("failed decoding tester object: %s", err)
	}

	return testerObject, nil
}

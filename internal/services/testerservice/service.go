package testerservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

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
		timeout: timeout,
		limiter: make(chan struct{}, 200), // max 200 parallel requests to tester service
	}
}

// GetObject fetches object from tester service
func (svc *ObjectService) GetObject(objectID uint) (model.TesterObject, error) {
	if objectID == 0 {
		return model.TesterObject{}, errors.New("no object ID provided")
	}

	select {
	case <-time.After(svc.timeout):
		return model.TesterObject{}, fmt.Errorf("get object timeout (%.0f secs)", svc.timeout.Seconds())
	case svc.limiter <- struct{}{}:
	}
	defer func() {
		<-svc.limiter
	}()

	url := fmt.Sprintf("http://tester:9010/objects/%d", objectID)

	response, err := httpClient.Get(url)
	if err != nil {
		return model.TesterObject{}, fmt.Errorf("get object details failed: %w", err)
	}
	defer response.Body.Close()

	var testerObject model.TesterObject
	if err := json.NewDecoder(response.Body).Decode(&testerObject); err != nil {
		return model.TesterObject{}, fmt.Errorf("failed decoding tester object: %w", err)
	}

	return testerObject, nil
}

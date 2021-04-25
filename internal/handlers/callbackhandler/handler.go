package callbackhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/evt/callback/internal/model"
)

// CallbackHandler is a callback handler.
type CallbackHandler struct {
	objectService ObjectService
	testerService TesterService
}

// New creates a new callback service.
func New(service ObjectService, testerService TesterService) *CallbackHandler {
	return &CallbackHandler{
		objectService: service,
		testerService: testerService,
	}
}

// Post handles POST /callback requests
func (h *CallbackHandler) Post(w http.ResponseWriter, r *http.Request) {
	// using default context as client waits for 5 seconds only which is not enough to complete
	defaultCtx := context.Background()

	var request model.CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %s", err), http.StatusBadRequest)
	}

	if len(request.ObjectIDs) == 0 {
		http.Error(w, "no object IDs provided", http.StatusBadRequest)
	}

	// this is a group of requests back to tester service for object details
	var wg sync.WaitGroup
	receivedObjects := make(chan model.TesterObject, len(request.ObjectIDs))

	for i := range request.ObjectIDs {
		wg.Add(1)

		objectID := request.ObjectIDs[i]

		// log.Printf("=> Next object ID: %d\n", objectID)

		go func() {
			defer wg.Done()

			object, err := h.testerService.GetObject(defaultCtx, objectID)
			if err != nil {
				log.Printf("[id: %d, total: %d] testerService.GetObject failed: %s\n", object.ID, len(request.ObjectIDs), err)

				return
			}

			// log.Printf("[id: %d, total: %d] testerService.GetObject passed (online=%t)\n", object.ID, len(request.ObjectIDs), object.Online)

			receivedObjects <- object
		}()
	}

	go func() {
		wg.Wait()
		close(receivedObjects)
	}()

	var totalUpdated, totalReceived int

	for object := range receivedObjects {
		totalReceived++

		if !object.Online {
			continue
		}

		h.objectService.UpdateObject(defaultCtx, &model.DBObject{
			ID: object.ID,
		})

		totalUpdated++
	}

	log.Printf("objects: received=%d, updated(=online) = %d\n", totalReceived, totalUpdated)

	w.Write([]byte("ok"))
}

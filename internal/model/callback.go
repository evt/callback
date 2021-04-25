package model

// CallbackRequest is a JSON request to /callback route.
type CallbackRequest struct {
	ObjectIDs []uint `json:"object_ids"`
}

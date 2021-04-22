package model

// TesterObject is a object returned by tester service.
type TesterObject struct {
	ID     uint `json:"id"`
	Online bool `json:"online"`
}

package model

import "fmt"

// Result is a an object returned by most dashboard api's
type Result struct {
	Status string

	// from dashboard api
	Message string
	Meta    string

	// from tyk api
	Key     string `json:"key"`
	Action  string `json:"action"`
	KeyHash string `json:"key_hash,omitempty"`
}

func (r *Result) String() string {
	return fmt.Sprintf("Status: %v, Message: %v", r.Status, r.Message)
}

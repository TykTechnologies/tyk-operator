package model

import "fmt"

// Result is a an object returned by most dashboard api's
type Result struct {
	Status     string
	StatusCode int

	// from dashboard api
	Message string
	Meta    string
	Errors  []string

	// from tyk api
	Key     string `json:"key"`
	Action  string `json:"action"`
	KeyHash string `json:"key_hash,omitempty"`
}

func (r *Result) String() string {
	msg := fmt.Sprintf("%v Status: %v HTTP %v", r.Message, r.Status, r.StatusCode)

	if r.Errors != nil {
		msg = fmt.Sprintf("%v: %v", msg, r.Errors)
	}

	return msg
}

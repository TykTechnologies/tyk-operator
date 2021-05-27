package model

// Result is a an object returned by most dashboard api's
type Result struct {
	Status string

	// from dashboard api
	Message string
	Meta    string

	//from tyk api
	Key     string `json:"key"`
	Action  string `json:"action"`
	KeyHash string `json:"key_hash,omitempty"`
}

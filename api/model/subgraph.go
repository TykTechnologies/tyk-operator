package model

// SubGraphSpec holds the configuration for a GraphQL federation subgraph.
type SubGraphSpec struct {
	SDL    string `json:"sdl"`
	Schema string `json:"schema"`
}

type SubGraphStatus struct {
	APIID string `json:"APIID"`
}

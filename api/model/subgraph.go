package model

type SubgraphConfig struct {
	SDL    string `json:"sdl"`
	Schema string `json:"schema"`
}

type SubGraphSpec struct {
	// Subgraph holds the configuration for a GraphQL federation subgraph.
	Subgraph SubgraphConfig `json:"subgraph,omitempty"`
}

type SubGraphStatus struct {
	APIID string `json:"APIID"`
}

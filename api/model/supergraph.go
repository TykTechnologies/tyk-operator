package model

// SuperGraphSpec defines the desired state of SuperGraph
type SuperGraphSpec struct {
	SubgraphRefs []Target `json:"subgraph_refs"`
	MergedSDL    string   `json:"merged_sdl,omitempty"`
	Schema       string   `json:"schema,omitempty"`
}

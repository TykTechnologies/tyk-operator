package model

// SuperGraphSpec defines the desired state of SuperGraph
type SuperGraphSpec struct {
	SubgraphsRefs []string `json:"subgraphs_refs"`
	MergedSDL     string   `json:"merged_sdl,omitempty"`
	Schema        string   `json:"schema,omitempty"`
}

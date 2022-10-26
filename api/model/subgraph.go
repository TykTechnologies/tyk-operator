package model

// SubGraphSpec holds the configuration for a GraphQL federation subgraph.
type SubGraphSpec struct {
	SDL    string `json:"sdl"`
	Schema string `json:"schema"`
}

type SubGraphStatus struct {
	// LinkedByAPI specifies the ID of the ApiDefinition CR that is linked to this particular SubGraph CR.
	// Please note that SubGraph CR can only be linked to one ApiDefinition CR that is created in the same
	// namespace as SubGraph CR.
	LinkedByAPI string `json:"linked_by_api,omitempty"`
}

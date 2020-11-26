package universal_client

type UniversalOrganization interface {
	GetID() string
}

func GetOrganizationID(c UniversalClient) string {
	return c.Organization().GetID()
}

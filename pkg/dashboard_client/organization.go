package dashboard_client

type Organization struct {
	*Client
}

func (o *Organization) GetID() string {
	return o.Env.Org
}

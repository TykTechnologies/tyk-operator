package dashboard_admin_client

import (
	"testing"
)

func TestClient_UserCreate(t *testing.T) {
	c := NewClient("http://localhost:3000", "12345", true)
	if err := c.UserCreate(CreateUserRequest{
		FirstName:    "Joe",
		LastName:     "Bloggs",
		EmailAddress: "joe@bloggs.com",
		Password:     "testing12345",
		Active:       true,
		OrgID:        "5f7f958a72a3b40001cb18dc",
	}); err != nil {
		t.Fatal(err.Error())
	}
}

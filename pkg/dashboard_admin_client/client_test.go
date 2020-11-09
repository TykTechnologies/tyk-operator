package dashboard_admin_client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/go-chi/chi"
)

func adminServerMock() *httptest.Server {
	r := chi.NewMux()
	r.Post(endpointOrganizations, mockOrganizationCreateHandler)
	r.Post(endpointUsers, mockUserCreateHandler)
	r.Post(fmt.Sprintf(endpointDashboardUserPasswordResetFormat, `{user_id}`), mockPasswordResetHandler)

	return httptest.NewServer(r)
}

func mockOrganizationCreateHandler(w http.ResponseWriter, r *http.Request) {
	svr := adminServerMock()
	defer svr.Close()

	resBody := CreateOrganizationResponse{
		Status:  "OK",
		Message: "ORG CREATED",
		Meta:    "MYORGID",
	}

	bodyBytes, _ := json.Marshal(resBody)
	w.Write(bodyBytes)
}

func mockUserCreateHandler(w http.ResponseWriter, r *http.Request) {
	resBody := CreateUserResponse{
		Status:  "OK",
		Message: "123",
		Meta: CreateUserMeta{
			CreateUserRequest: CreateUserRequest{
				FirstName:    "Joe",
				LastName:     "Bloggs",
				EmailAddress: "joe@bloggs.com",
				Active:       true,
				OrgID:        "5f7f958a72a3b40001cb18dc",
			},
			ID:        "54321",
			AccessKey: "abcde",
			UserPermissions: struct {
				IsAdmin       string `json:"IsAdmin"`
				ResetPassword string `json:"ResetPassword"`
			}{},
		},
	}

	resBodyBytes, _ := json.Marshal(resBody)
	w.Write(resBodyBytes)
}

func mockPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func TestClient_OrganizationCreate(t *testing.T) {
	svr := adminServerMock()
	defer svr.Close()

	c := NewClient(svr.URL, "12345", true)
	createdOrgID, err := c.OrganizationCreate(&v1alpha1.OrganizationSpec{
		OwnerName:    "",
		OwnerSlug:    "",
		CNAMEEnabled: false,
		CNAME:        "",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(createdOrgID)
}

func TestClient_UserCreate(t *testing.T) {
	svr := adminServerMock()
	defer svr.Close()

	c := NewClient(svr.URL, "12345", true)
	if _, err := c.UserAdminCreate(CreateUserRequest{
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

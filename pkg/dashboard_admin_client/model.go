package dashboard_admin_client

import (
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type OrganizationsResponse struct {
	Organizations []v1alpha1.OrganizationSpec `json:"organisations"`
}

type CreateOrganizationResponse struct {
	Status  string `json:"Status"`
	Message string `json:"message"`
	Meta    string `json:"meta"`
}

type CreateUserRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	EmailAddress string `json:"email_address"`
	Password     string `json:"password"`
	Active       bool   `json:"active"`
	OrgID        string `json:"org_id"`
}

//{"Status":"OK","Message":"63da42beed3840bd4cfe69fe3705c9b9","Meta":{"api_model":{},"first_name":"Ahmet","last_name":"Soormally","email_address":"ahmet@tyk.io","org_id":"5f7f958a72a3b40001cb18dc","active":true,"id":"5f7f98427fea06e1b637b2b9","access_key":"63da42beed3840bd4cfe69fe3705c9b9","user_permissions":{"IsAdmin":"admin","ResetPassword":"admin"},"group_id":"","password_max_days":0,"password_updated":"0001-01-01T00:00:00Z","PWHistory":[],"last_login_date":"0001-01-01T00:00:00Z","created_at":"2020-10-08T22:52:50.442Z"}}
type CreateUserResponse struct {
	Status  string         `json:"Status"`
	Message string         `json:"Message"`
	Meta    CreateUserMeta `json:"Meta"`
}

type CreateUserMeta struct {
	CreateUserRequest
	ID              string `json:"id"`
	AccessKey       string `json:"access_key"`
	UserPermissions struct {
		IsAdmin       string `json:"IsAdmin"`
		ResetPassword string `json:"ResetPassword"`
	} `json:"user_permissions"`
	GroupID string `json:"group_id"`
}

type SetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

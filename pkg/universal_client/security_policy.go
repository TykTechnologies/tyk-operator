package universal_client

import (
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

type UniversalSecurityPolicy interface {
	All() ([]tykv1alpha1.SecurityPolicySpec, error)
	// Get retruns the policy with the given id.
	Get(id string) (*tykv1alpha1.SecurityPolicySpec, error)
	// Create creates a new def and updates id and other fields. It is up to the
	// caller to update any fields that will be set after the policy has been
	// created for instance _id
	Create(def *tykv1alpha1.SecurityPolicySpec) error
	// Update this will update an existing policy
	Update(def *tykv1alpha1.SecurityPolicySpec) error
	//Delete deletes policy id id
	Delete(id string) error
}

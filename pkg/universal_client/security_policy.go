package universal_client

import (
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/pkg/errors"
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

var (
	PolicyCollisionError = errors.New("policy already exists")
	PolicyNotFoundError  = errors.New("policy not found")
)

func applyDefaults(spec *tykv1alpha1.SecurityPolicySpec) {
	if spec.Rate == 0 || spec.Per == 0 {
		spec.Rate = -1
		spec.Per = -1
	}
	if spec.ThrottleInterval == 0 || spec.ThrottleRetryLimit == 0 {
		spec.ThrottleInterval = -1
		spec.ThrottleRetryLimit = -1
	}
	if spec.QuotaMax == 0 || spec.QuotaRenewalRate == 0 {
		spec.QuotaMax = -1
		spec.QuotaRenewalRate = -1
	}
	if spec.MaxQueryDepth == 0 {
		spec.MaxQueryDepth = -1
	}
}

func CreateOrUpdatePolicy(c UniversalClient, spec *tykv1alpha1.SecurityPolicySpec) (*tykv1alpha1.SecurityPolicySpec, error) {
	var err error

	pol, err := c.SecurityPolicy().Get(spec.ID)
	if err != nil {
		// should return "nil, http.401" if policy doesn't exist
		if !errors.Is(err, PolicyNotFoundError) {
			return nil, errors.Wrap(err, "Unable to communicate with Client")
		}
	}

	applyDefaults(spec)

	if pol == nil {
		// Create
		_, err := c.SecurityPolicy().Create(spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create policy")
		}

	} else {
		// Update
		spec.MID = pol.MID
		spec.OrgID = pol.OrgID
		err = c.SecurityPolicy().Update(spec)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update policy")
		}
	}

	_ = c.HotReload()

	return pol, nil
}

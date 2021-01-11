/*


Licensed under the Mozilla Public License (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.mozilla.org/en-US/MPL/2.0/

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"os"
	"strings"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var securitypolicylog = logf.Log.WithName("securitypolicy-resource")

func (r *SecurityPolicy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-tyk-tyk-io-v1alpha1-securitypolicy,mutating=true,failurePolicy=fail,groups=tyk.tyk.io,resources=securitypolicies,verbs=create;update,versions=v1alpha1,name=msecuritypolicy.kb.io,sideEffects=None,admissionReviewVersions=v1,webhookVersions=v1

var _ webhook.Defaulter = &SecurityPolicy{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SecurityPolicy) Default() {
	securitypolicylog.Info("default", "name", r.Name)
	spec := r.Spec
	spec.Rate = -1
	spec.Per = -1
	spec.ThrottleInterval = -1
	spec.ThrottleRetryLimit = -1
	spec.QuotaMax = -1
	spec.QuotaRenewalRate = -1
	spec.OrgID = strings.TrimSpace(os.Getenv(environmet.TykORG))
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-tyk-tyk-io-v1alpha1-securitypolicy,mutating=false,failurePolicy=fail,groups=tyk.tyk.io,resources=securitypolicies,versions=v1alpha1,name=vsecuritypolicy.kb.io,sideEffects=None,admissionReviewVersions=v1,webhookVersions=v1

var _ webhook.Validator = &SecurityPolicy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SecurityPolicy) ValidateCreate() error {
	securitypolicylog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SecurityPolicy) ValidateUpdate(old runtime.Object) error {
	securitypolicylog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SecurityPolicy) ValidateDelete() error {
	securitypolicylog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

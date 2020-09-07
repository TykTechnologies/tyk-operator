/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var apidefinitionlog = logf.Log.WithName("apidefinition-resource")

func (in *ApiDefinition) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-tyk-tyk-io-v1-apidefinition,mutating=true,failurePolicy=fail,groups=tyk.tyk.io,resources=apidefinitions,verbs=create;update,versions=v1,name=mapidefinition.kb.io

var _ webhook.Defaulter = &ApiDefinition{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *ApiDefinition) Default() {
	apidefinitionlog.Info("default", "name", in.Name)

	if len(in.Spec.VersionData.Versions) == 0 {
		// TODO: this prob belongs in a mutating webhook
		defaultVersionData := VersionData{
			NotVersioned:   true,
			DefaultVersion: "Default",
			Versions: map[string]VersionInfo{
				"Default": {
					Name:                        "Default",
					Expires:                     "",
					Paths:                       VersionInfoPaths{},
					UseExtendedPaths:            false,
					ExtendedPaths:               ExtendedPathsSet{},
					GlobalHeaders:               nil,
					GlobalHeadersRemove:         nil,
					GlobalResponseHeaders:       nil,
					GlobalResponseHeadersRemove: nil,
					IgnoreEndpointCase:          false,
					GlobalSizeLimit:             0,
					OverrideTarget:              "",
				},
			},
		}

		in.Spec.VersionData = defaultVersionData
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-tyk-tyk-io-v1-apidefinition,mutating=false,failurePolicy=fail,groups=tyk.tyk.io,resources=apidefinitions,versions=v1,name=vapidefinition.kb.io

var _ webhook.Validator = &ApiDefinition{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateCreate() error {
	apidefinitionlog.Info("validate create", "name", in.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateUpdate(old runtime.Object) error {
	apidefinitionlog.Info("validate update", "name", in.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateDelete() error {
	apidefinitionlog.Info("validate delete", "name", in.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

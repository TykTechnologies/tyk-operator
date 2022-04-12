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
	"net/url"

	"github.com/TykTechnologies/tyk-operator/api/model"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var (
	apidefinitionlog = logf.Log.WithName("apidefinition-resource")
	ErrEmptyValue    = "can't be empty"
)

func (in *ApiDefinition) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-tyk-tyk-io-v1alpha1-apidefinition,mutating=true,failurePolicy=fail,groups=tyk.tyk.io,resources=apidefinitions,verbs=create;update,versions=v1alpha1,name=mapidefinition.kb.io,sideEffects=None,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &ApiDefinition{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *ApiDefinition) Default() {
	apidefinitionlog.Info("default", "name", in.Name)

	if len(in.Spec.VersionData.Versions) == 0 {
		in.Spec.VersionData = model.VersionData{
			NotVersioned:   true,
			DefaultVersion: "Default",
			Versions: map[string]model.VersionInfo{
				"Default": {
					Name:             "Default",
					UseExtendedPaths: false,
				},
			},
		}
	}

	if in.Spec.UseStandardAuth {
		if in.Spec.AuthConfigs == nil {
			in.Spec.AuthConfigs = make(map[string]model.AuthConfig)
		}

		if _, ok := in.Spec.AuthConfigs["authToken"]; !ok {
			apidefinitionlog.Info("applying default auth_config as not set & use_standard_auth enabled")

			in.Spec.AuthConfigs["authToken"] = model.AuthConfig{
				AuthHeaderName: "Authorization",
			}
		}
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-tyk-tyk-io-v1alpha1-apidefinition,mutating=false,failurePolicy=fail,groups=tyk.tyk.io,resources=apidefinitions,versions=v1alpha1,name=vapidefinition.kb.io,sideEffects=None,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &ApiDefinition{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateCreate() error {
	apidefinitionlog.Info("validate create", "name", in.Name)
	return in.validate()
}

func path(n ...string) *field.Path {
	x := field.NewPath("spec")

	for _, v := range n {
		x = x.Child(v)
	}

	return x
}

func (in *ApiDefinition) validate() error {
	var all field.ErrorList

	spec := in.Spec

	// auth
	if spec.UseKeylessAccess {
		if spec.UseStandardAuth {
			all = append(all,
				field.Forbidden(path("use_standard_auth"), "use_keyless_access and use_standard_auth cannot be set together"),
			)
		}
	} else {
		if spec.UseStandardAuth {
			if len(spec.AuthConfigs) > 0 {
				_, ok := spec.AuthConfigs["authToken"]
				if !ok {
					all = append(all,
						field.NotFound(path("auth_configs", "authToken"), nil),
					)
				}
			} else {
				all = append(all,
					field.NotFound(path("auth_configs"), nil),
				)
			}
		}
	}

	// graphql
	if spec.GraphQL != nil {
		if spec.GraphQL.Enabled && spec.GraphQL.ExecutionMode == "executionEngine" {
			for _, typeFieldConfig := range spec.GraphQL.TypeFieldConfigurations {
				switch typeFieldConfig.DataSource.Kind {
				case "HTTPJsonDataSource", "GraphQLDataSource":
					src := typeFieldConfig.DataSource
					if src.Config.URL == "" {
						all = append(all,
							field.Required(path("graphql", "type_field_configurations", "data_source", "url"),
								ErrEmptyValue,
							),
						)
					} else {
						_, err := url.Parse(src.Config.URL)
						if err != nil {
							all = append(all,
								field.Invalid(path("graphql", "type_field_configurations", "data_source", "url"),
									src.Config.URL,
									err.Error(),
								),
							)
						}
					}

					if src.Config.Method == "" {
						all = append(all,
							field.Required(path("graphql", "type_field_configurations", "data_source", "method"),
								ErrEmptyValue,
							),
						)
					}
				default:
					all = append(all,
						field.Invalid(path("graphql", "type_field_configurations", "data_source", "kind"),
							typeFieldConfig.DataSource.Kind,
							"invalid data source kind type",
						),
					)
				}
			}
		}
	}

	// proxy
	if a := in.validateTarget(); len(a) > 0 {
		all = append(all, a...)
	}

	if len(all) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{
			Group: "tyk.tyk.io",
			Kind:  "ApiDefinition",
		},
		in.Name, all,
	)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateUpdate(old runtime.Object) error {
	apidefinitionlog.Info("validate update", "name", in.Name)
	return in.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateDelete() error {
	apidefinitionlog.Info("validate delete", "name", in.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (in *ApiDefinition) validateTarget() field.ErrorList {
	var all field.ErrorList
	// TargetURL is only allowed to be an empty string when we have GraphQL
	// API, or we are targeting internal API by its name and namespace.
	if in.Spec.Proxy.TargetURL == "" {
		if in.Spec.GraphQL != nil {
			if !in.Spec.GraphQL.Enabled {
				all = append(all,
					field.Required(path("proxy", "target_url"),
						ErrEmptyValue,
					),
				)
			}
		} else if in.Spec.Proxy.TargetInternal == nil {
			all = append(all,
				field.Required(path("proxy", "target_url"),
					ErrEmptyValue,
				),
			)
		}
	}

	for _, v := range in.Spec.VersionData.Versions {
		if v.ExtendedPaths != nil {
			for _, u := range v.ExtendedPaths.URLRewrite {
				if u.RewriteTo == "" && u.RewriteToInternal == nil && len(u.Triggers) == 0 {
					all = append(all,
						field.Required(path("version_data", "versions", v.Name, "extended_paths", "url_rewrites", "rewrite_to"),
							ErrEmptyValue,
						),
					)
				}

				for _, t := range u.Triggers {
					if t.RewriteTo == "" && t.RewriteToInternal == nil {
						all = append(all,
							field.Required(path("version_data", "versions", v.Name, "extended_paths", "url_rewrites", "triggers", "rewrite_to"),
								ErrEmptyValue,
							),
						)
					}
				}
			}
		}
	}

	return all
}

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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
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

// +kubebuilder:webhook:path=/mutate-tyk-tyk-io-v1alpha1-apidefinition,mutating=true,failurePolicy=fail,groups=tyk.tyk.io,resources=apidefinitions,verbs=create;update,versions=v1alpha1,name=mapidefinition.kb.io,sideEffects=None

var _ webhook.Defaulter = &ApiDefinition{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *ApiDefinition) Default() {
	apidefinitionlog.Info("default", "name", in.Name)

	// We disable tracking by default
	if in.Spec.DoNotTrack == nil {
		in.Spec.DoNotTrack = pointer.BoolPtr(true)
	}

	if len(in.Spec.VersionData.Versions) == 0 {
		defaultVersionData := VersionData{
			NotVersioned:   true,
			DefaultVersion: "Default",
			Versions: map[string]VersionInfo{
				"Default": {
					Name:    "Default",
					Expires: "",
					Paths: VersionInfoPaths{
						Ignored:   []string{},
						WhiteList: []string{},
						BlackList: []string{},
					},
					UseExtendedPaths:            false,
					ExtendedPaths:               ExtendedPathsSet{},
					GlobalHeaders:               nil,
					GlobalHeadersRemove:         nil,
					GlobalResponseHeaders:       nil,
					GlobalResponseHeadersRemove: nil,
					IgnoreEndpointCase:          false,
					GlobalSizeLimit:             0,
				},
			},
		}

		in.Spec.VersionData = defaultVersionData
	}

	if in.Spec.UseStandardAuth {
		if in.Spec.AuthConfigs == nil {
			in.Spec.AuthConfigs = make(map[string]AuthConfig)
		}
		if _, ok := in.Spec.AuthConfigs["authToken"]; !ok {
			apidefinitionlog.Info("applying default auth_config as not set & use_standard_auth enabled")
			in.Spec.AuthConfigs["authToken"] = AuthConfig{
				AuthHeaderName: "Authorization",
			}
		}
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-tyk-tyk-io-v1alpha1-apidefinition,mutating=false,failurePolicy=fail,groups=tyk.tyk.io,resources=apidefinitions,versions=v1alpha1,name=vapidefinition.kb.io,sideEffects=None

var _ webhook.Validator = &ApiDefinition{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *ApiDefinition) ValidateCreate() error {
	apidefinitionlog.Info("validate create", "name", in.Name)
	return in.validate()
}

func (in *ApiDefinition) validate() error {
	var all field.ErrorList
	var _ APIDefinitionSpec

	path := func(n ...string) *field.Path {
		x := field.NewPath("spec")
		for _, v := range n {
			x = x.Child(v)
		}
		return x
	}

	spec := in.Spec
	// protocol
	switch spec.Protocol {
	case "", "h2c", "tcp", "tls", "http", "https":
	default:
		all = append(all,
			field.NotSupported(path("protocol"), spec.Protocol, []string{"", "h2c", "tcp", "tls", "http", "https"}),
		)
	}

	// auth
	if spec.UseKeylessAccess {
		if spec.UseStandardAuth {
			all = append(all,
				field.Forbidden(path("use_standard_auth"), "use_keyless_access & use_standard_auth"),
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
							field.Invalid(path("graphql", "type_field_configurations", "data_source", "url"),
								"",
								"can't be emptry",
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
							field.Invalid(path("graphql", "type_field_configurations", "data_source", "method"),
								string(src.Config.Method),
								"can't be emptry",
							),
						)
					}
				}
			}
		}
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

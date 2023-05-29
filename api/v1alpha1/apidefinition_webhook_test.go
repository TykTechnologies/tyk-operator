package v1alpha1

import (
	"net/http"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/matryer/is"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestApiDefinition_Default(t *testing.T) {
	useStandardAuth := true

	in := ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{UseStandardAuth: &useStandardAuth},
		},
	}
	in.Default()

	if !in.Spec.VersionData.NotVersioned {
		t.Fatal("expected the api to not be versioned")
	}

	if in.Spec.VersionData.DefaultVersion != "Default" {
		t.Fatal("expected default version to be Default")
	}

	if len(in.Spec.VersionData.Versions) == 0 {
		t.Fatal("expected default version to be applied")
	}

	authConf, ok := in.Spec.AuthConfigs["authToken"]
	if !ok {
		t.Fatal("we used standard auth, so the authToken config must be set")
	}

	if authConf.AuthHeaderName != "Authorization" {
		t.Fatal("expected the authConf.AuthHeaderName to be Authorization, Got", authConf.AuthHeaderName)
	}
}

func TestApiDefinition_validateTarget(t *testing.T) {
	rewriteTo := ""
	namespace := "resource-ns"

	invalidRewriteApiDef := ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Proxy: model.Proxy{TargetURL: "/test"},
				VersionData: model.VersionData{
					Versions: map[string]model.VersionInfo{
						"Default": {
							ExtendedPaths: &model.ExtendedPathsSet{
								URLRewrite: []model.URLRewriteMeta{{RewriteTo: &rewriteTo}},
							},
						},
					},
				},
			},
		},
	}

	invalidTriggerApiDef := ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Proxy: model.Proxy{TargetURL: "/test"},
				VersionData: model.VersionData{
					Versions: map[string]model.VersionInfo{
						"Default": {
							ExtendedPaths: &model.ExtendedPathsSet{
								URLRewrite: []model.URLRewriteMeta{
									{RewriteTo: &rewriteTo, Triggers: []model.RoutingTrigger{{}}},
								},
							},
						},
					},
				},
			},
		},
	}

	missingTargetURLErr := field.Error{
		Type:   field.ErrorTypeRequired,
		Field:  path("proxy", "target_url").String(),
		Detail: ErrEmptyValue,
	}

	cases := []struct {
		name        string
		apiDef      ApiDefinition
		expectedErr *field.Error
	}{
		{
			name: "valid ApiDefinition",
			apiDef: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						Proxy: model.Proxy{TargetURL: "/test"},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "valid ApiDefinition for UDG",
			apiDef: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						Proxy:   model.Proxy{TargetURL: ""},
						GraphQL: &model.GraphQLConfig{Enabled: true},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "valid ApiDefinition, targeting internal API",
			apiDef: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						Proxy: model.Proxy{TargetURL: "", TargetInternal: &model.TargetInternal{
							Target: model.Target{Name: "resource-name", Namespace: &namespace},
						}},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:        "invalid ApiDefinition, missing proxy.target_url",
			apiDef:      ApiDefinition{},
			expectedErr: &missingTargetURLErr,
		},
		{
			name: "invalid ApiDefinition, empty proxy.target_url",
			apiDef: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						Proxy: model.Proxy{TargetURL: ""},
					},
				},
			},
			expectedErr: &missingTargetURLErr,
		},
		{
			name: "invalid ApiDefinition, missing proxy.target_url in GraphQL",
			apiDef: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						GraphQL: &model.GraphQLConfig{Enabled: false},
					},
				},
			},
			expectedErr: &missingTargetURLErr,
		},
		{
			name:   "invalid ApiDefinition, missing URLRewrite in ExtendedPaths",
			apiDef: invalidRewriteApiDef,
			expectedErr: &field.Error{
				Type:   field.ErrorTypeRequired,
				Field:  path("version_data", "versions", "", "extended_paths", "url_rewrites", "rewrite_to").String(),
				Detail: ErrEmptyValue,
			},
		},
		{
			name:   "invalid ApiDefinition, missing URLRewrite in url_rewrites.triggers",
			apiDef: invalidTriggerApiDef,
			expectedErr: &field.Error{
				Type: field.ErrorTypeRequired,
				Field: path(
					"version_data", "versions", "", "extended_paths", "url_rewrites", "triggers", "rewrite_to",
				).String(),
				Detail: ErrEmptyValue,
			},
		},
	}

	for _, tc := range cases {
		errs := tc.apiDef.validateTarget()

		// If validateTarget returns errors even we do not expect any errors, return failure.
		if len(errs) != 0 && tc.expectedErr == nil {
			t.Errorf("%s: unexpected number of errors occured, expected 0, got %d", tc.name, len(errs))
		}

		if tc.expectedErr != nil && !hasError(errs, tc.expectedErr.Error()) {
			t.Errorf("%s: got %v, want %v", tc.name, errs, tc.expectedErr)
		}
	}
}

func hasError(errs field.ErrorList, needle string) bool {
	for _, curr := range errs {
		if curr.Error() == needle {
			return true
		}
	}

	return false
}

func TestApiDefinition_Validate_Auth(t *testing.T) {
	eval := is.New(t)

	useKeyLess := true

	tests := map[string]struct {
		ApiDefinition ApiDefinition
		ErrCause      field.ErrorType
	}{
		"set both keyless and auth type": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseKeylessAccess: &useKeyLess,
						UseStandardAuth:  &useKeyLess,
						Proxy:            model.Proxy{TargetURL: "/test"},
					},
				},
			},
			ErrCause: field.ErrorTypeForbidden,
		},
		"set keyless auth type": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseKeylessAccess: &useKeyLess,
						Proxy:            model.Proxy{TargetURL: "/test"},
					},
				},
			},
		},
		"set standard auth without auth details": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseStandardAuth: &useKeyLess,
						Proxy:           model.Proxy{TargetURL: "/test"},
					},
				},
			},
			ErrCause: field.ErrorTypeNotFound,
		},
		"set standard auth without authToken details": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseStandardAuth: &useKeyLess,
						AuthConfigs: map[string]model.AuthConfig{
							"random": {
								AuthHeaderName: "Authorization",
							},
						},
						Proxy: model.Proxy{TargetURL: "/test"},
					},
				},
			},
			ErrCause: field.ErrorTypeNotFound,
		},
		"set standard auth with authToken details": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseStandardAuth: &useKeyLess,
						AuthConfigs: map[string]model.AuthConfig{
							"authToken": {
								AuthHeaderName: "Authorization",
							},
						},
						Proxy: model.Proxy{TargetURL: "/test"},
					},
				},
			},
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			if err := tc.ApiDefinition.validate(); err != nil {
				eval.True(apierrors.IsInvalid(err))
				eval.True(apierrors.HasStatusCause(err, metav1.CauseType(tc.ErrCause)))
			}
		})
	}
}

func TestApiDefinition_Validate_GraphQLDataSource(t *testing.T) {
	eval := is.New(t)
	useKeyLess := true

	tests := map[string]struct {
		ApiDefinition ApiDefinition
		ErrCause      field.ErrorType
	}{
		"empty data source kind": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseKeylessAccess: &useKeyLess,
						Proxy:            model.Proxy{TargetURL: "/test"},
						GraphQL: &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: "executionEngine",
							TypeFieldConfigurations: []model.TypeFieldConfiguration{
								{
									DataSource: model.SourceConfig{
										Kind: "",
									},
								},
							},
						},
					},
				},
			},
			ErrCause: field.ErrorTypeInvalid,
		},
		"invalid data source kind": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseKeylessAccess: &useKeyLess,
						Proxy:            model.Proxy{TargetURL: "/test"},
						GraphQL: &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: "executionEngine",
							TypeFieldConfigurations: []model.TypeFieldConfiguration{
								{
									DataSource: model.SourceConfig{
										Kind: "invalid",
									},
								},
							},
						},
					},
				},
			},
			ErrCause: field.ErrorTypeInvalid,
		},
		"valid data source with empty URL": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseKeylessAccess: &useKeyLess,
						Proxy:            model.Proxy{TargetURL: "/test"},
						GraphQL: &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: "executionEngine",
							TypeFieldConfigurations: []model.TypeFieldConfiguration{
								{
									DataSource: model.SourceConfig{
										Kind:   "HTTPJsonDataSource",
										Config: model.DataSourceConfig{},
									},
								},
							},
						},
					},
				},
			},
			ErrCause: field.ErrorTypeRequired,
		},
		"valid data source with invalid URL": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseKeylessAccess: &useKeyLess,
						Proxy:            model.Proxy{TargetURL: "/test"},
						GraphQL: &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: "executionEngine",
							TypeFieldConfigurations: []model.TypeFieldConfiguration{
								{
									DataSource: model.SourceConfig{
										Kind: "HTTPJsonDataSource",
										Config: model.DataSourceConfig{
											URL:    "hi/\there?",
											Method: http.MethodGet,
										},
									},
								},
							},
						},
					},
				},
			},
			ErrCause: field.ErrorTypeInvalid,
		},
		"valid data source with empty method": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseKeylessAccess: &useKeyLess,
						Proxy:            model.Proxy{TargetURL: "/test"},
						GraphQL: &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: "executionEngine",
							TypeFieldConfigurations: []model.TypeFieldConfiguration{
								{
									DataSource: model.SourceConfig{
										Kind: "HTTPJsonDataSource",
										Config: model.DataSourceConfig{
											URL: "http://httpbin.org",
										},
									},
								},
							},
						},
					},
				},
			},
			ErrCause: field.ErrorTypeRequired,
		},
		"valid api with HTTP DataSource": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseKeylessAccess: &useKeyLess,
						Proxy:            model.Proxy{TargetURL: "/test"},
						GraphQL: &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: "executionEngine",
							TypeFieldConfigurations: []model.TypeFieldConfiguration{
								{
									DataSource: model.SourceConfig{
										Kind: "HTTPJsonDataSource",
										Config: model.DataSourceConfig{
											URL:    "http://httpbin.org",
											Method: http.MethodGet,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"valid api with GraphQL DataSource": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						UseKeylessAccess: &useKeyLess,
						Proxy:            model.Proxy{TargetURL: "/test"},
						GraphQL: &model.GraphQLConfig{
							Enabled:       true,
							ExecutionMode: "executionEngine",
							TypeFieldConfigurations: []model.TypeFieldConfiguration{
								{
									DataSource: model.SourceConfig{
										Kind: "GraphQLDataSource",
										Config: model.DataSourceConfig{
											URL:    "http://httpbin.org",
											Method: http.MethodGet,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			if err := tc.ApiDefinition.validate(); err != nil {
				statusErr, ok := err.(*apierrors.StatusError)
				if !ok {
					t.Fatal("invalid error type")
				}
				eval.True(apierrors.IsInvalid(err))

				t.Log(statusErr.Status().Details.Causes)

				eval.True(apierrors.HasStatusCause(err, metav1.CauseType(tc.ErrCause)))
			}
		})
	}
}

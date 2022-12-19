package v1alpha1

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/TykTechnologies/graphql-go-tools/pkg/execution/datasource"
	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/matryer/is"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestApiDefinition_Default(t *testing.T) {
	in := ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{APIDefinition: apidef.APIDefinition{UseStandardAuth: true}},
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
	invalidRewriteApiDef := ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
				VersionData: model.VersionData{
					Versions: map[string]model.VersionInfo{
						"Default": {
							ExtendedPaths: &model.ExtendedPathsSet{
								URLRewrite: []model.URLRewriteMeta{{URLRewriteMeta: apidef.URLRewriteMeta{RewriteTo: ""}}},
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
				Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
				VersionData: model.VersionData{
					Versions: map[string]model.VersionInfo{
						"Default": {
							ExtendedPaths: &model.ExtendedPathsSet{
								URLRewrite: []model.URLRewriteMeta{
									{URLRewriteMeta: apidef.URLRewriteMeta{RewriteTo: ""}, Triggers: []model.RoutingTrigger{{}}},
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
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
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
						Proxy:         model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: ""}},
						APIDefinition: apidef.APIDefinition{GraphQL: apidef.GraphQLConfig{Enabled: true}},
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
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: ""}, TargetInternal: &model.TargetInternal{
							Target: model.Target{Name: "resource-name", Namespace: "resource-ns"},
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
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: ""}},
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
						APIDefinition: apidef.APIDefinition{
							GraphQL: apidef.GraphQLConfig{Enabled: false},
						}},
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
	is := is.New(t)

	tests := map[string]struct {
		ApiDefinition ApiDefinition
		ErrCause      field.ErrorType
	}{
		"set both keyless and auth type": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseKeylessAccess: true,
							UseStandardAuth:  true,
						},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
			ErrCause: field.ErrorTypeForbidden,
		},
		"set keyless auth type": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseKeylessAccess: true},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
		},
		"set standard auth without auth details": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseStandardAuth: true,
						},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
			ErrCause: field.ErrorTypeNotFound,
		},
		"set standard auth without authToken details": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseStandardAuth: true,
							AuthConfigs: map[string]apidef.AuthConfig{
								"random": {
									AuthHeaderName: "Authorization",
								},
							}},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
			ErrCause: field.ErrorTypeNotFound,
		},
		"set standard auth with authToken details": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseStandardAuth: true,
							AuthConfigs: map[string]apidef.AuthConfig{
								"authToken": {
									AuthHeaderName: "Authorization",
								},
							}},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
		},
	}

	for n, tc := range tests {
		t.Run(n, func(t *testing.T) {
			if err := tc.ApiDefinition.validate(); err != nil {
				is.True(apierrors.IsInvalid(err))
				is.True(apierrors.HasStatusCause(err, metav1.CauseType(tc.ErrCause)))
			}
		})
	}
}

func TestApiDefinition_Validate_GraphQLDataSource(t *testing.T) {
	is := is.New(t)

	tests := map[string]struct {
		ApiDefinition ApiDefinition
		ErrCause      field.ErrorType
	}{
		"empty data source kind": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseKeylessAccess: true,
							GraphQL: apidef.GraphQLConfig{
								Enabled:       true,
								ExecutionMode: "executionEngine",
								TypeFieldConfigurations: []datasource.TypeFieldConfiguration{
									{
										DataSource: datasource.SourceConfig{
											Name: "",
										},
									},
								},
							}},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
			ErrCause: field.ErrorTypeInvalid,
		},
		"invalid data source kind": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseKeylessAccess: true,
							GraphQL: apidef.GraphQLConfig{
								Enabled:       true,
								ExecutionMode: "executionEngine",
								TypeFieldConfigurations: []datasource.TypeFieldConfiguration{
									{
										DataSource: datasource.SourceConfig{
											Name: "invalid",
										},
									},
								},
							}},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
			ErrCause: field.ErrorTypeInvalid,
		},
		"valid data source with empty URL": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseKeylessAccess: true,
							GraphQL: apidef.GraphQLConfig{
								Enabled:       true,
								ExecutionMode: "executionEngine",
								TypeFieldConfigurations: []datasource.TypeFieldConfiguration{
									{
										DataSource: datasource.SourceConfig{
											Name: "HTTPJsonDataSource",
										},
									},
								},
							}},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
			ErrCause: field.ErrorTypeRequired,
		},
		"valid data source with invalid URL": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseKeylessAccess: true,
							GraphQL: apidef.GraphQLConfig{
								Enabled:       true,
								ExecutionMode: "executionEngine",
								TypeFieldConfigurations: []datasource.TypeFieldConfiguration{
									{
										DataSource: datasource.SourceConfig{
											Name: "HTTPJsonDataSource",
											Config: func() json.RawMessage {
												data, _ := json.Marshal(datasource.HttpJsonDataSourceConfig{
													URL: "hi/\there?",
													Method: func() *string {
														method := http.MethodGet
														return &method
													}(),
												})
												return data
											}(),
										},
									},
								},
							}},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
			ErrCause: field.ErrorTypeInvalid,
		},
		"valid data source with empty method": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseKeylessAccess: true,
							GraphQL: apidef.GraphQLConfig{
								Enabled:       true,
								ExecutionMode: "executionEngine",
								TypeFieldConfigurations: []datasource.TypeFieldConfiguration{
									{
										DataSource: datasource.SourceConfig{
											Name: "HTTPJsonDataSource",
											Config: func() json.RawMessage {
												data, _ := json.Marshal(datasource.HttpJsonDataSourceConfig{
													URL: "http://httpbin.org",
												})

												return data
											}(),
										},
									},
								},
							}},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
			ErrCause: field.ErrorTypeRequired,
		},
		"valid api with HTTP DataSource": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseKeylessAccess: true,
							GraphQL: apidef.GraphQLConfig{
								Enabled:       true,
								ExecutionMode: "executionEngine",
								TypeFieldConfigurations: []datasource.TypeFieldConfiguration{
									{
										DataSource: datasource.SourceConfig{
											Name: "HTTPJsonDataSource",
											Config: func() json.RawMessage {
												data, _ := json.Marshal(datasource.HttpJsonDataSourceConfig{
													URL: "http://httpbin.org",
													Method: func() *string {
														method := http.MethodGet
														return &method
													}(),
												})

												return data
											}(),
										},
									},
								},
							}},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
					},
				},
			},
		},
		"valid api with GraphQL DataSource": {
			ApiDefinition: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						APIDefinition: apidef.APIDefinition{
							UseKeylessAccess: true,
							GraphQL: apidef.GraphQLConfig{
								Enabled:       true,
								ExecutionMode: "executionEngine",
								TypeFieldConfigurations: []datasource.TypeFieldConfiguration{
									{
										DataSource: datasource.SourceConfig{
											Name: "GraphQLDataSource",
											Config: func() json.RawMessage {
												data, _ := json.Marshal(datasource.GraphQLDataSourceConfig{
													URL: "http://httpbin.org",
													Method: func() *string {
														method := http.MethodGet
														return &method
													}(),
												})

												return data
											}(),
										},
									},
								},
							}},
						Proxy: model.ProxyConfig{ProxyConfig: apidef.ProxyConfig{TargetURL: "/test"}},
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
				is.True(apierrors.IsInvalid(err))

				t.Log(statusErr.Status().Details.Causes)

				is.True(apierrors.HasStatusCause(err, metav1.CauseType(tc.ErrCause)))
			}
		})
	}
}

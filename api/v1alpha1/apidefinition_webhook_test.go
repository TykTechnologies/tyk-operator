package v1alpha1

import (
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
)

func TestApiDefinition_Default(t *testing.T) {
	in := ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{UseStandardAuth: true},
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

func TestApiDefinition_Default_DoNotTrack(t *testing.T) {
	in := ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				UseStandardAuth: true,
				DoNotTrack:      pointer.BoolPtr(true),
			},
		},
	}
	in.Default()

	if *in.Spec.DoNotTrack != true {
		t.Fatalf("expected DoNotTrack to be true by default, got %v", *in.Spec.DoNotTrack)
	}

	in = ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				UseStandardAuth: true,
				DoNotTrack:      nil,
			},
		},
	}
	in.Default()

	if *in.Spec.DoNotTrack != true {
		t.Fatalf("expected DoNotTrack to be true as explicitly set, got %v", *in.Spec.DoNotTrack)
	}

	in = ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				UseStandardAuth: true,
				DoNotTrack:      pointer.BoolPtr(false),
			},
		},
	}
	in.Default()

	if *in.Spec.DoNotTrack != false {
		t.Fatalf("expected DoNotTrack to be false as explicitly set, got %v", *in.Spec.DoNotTrack)
	}
}

func TestApiDefinition_validateTarget(t *testing.T) {
	invalidRewriteApiDef := ApiDefinition{
		Spec: APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				Proxy: model.Proxy{TargetURL: "/test"},
				VersionData: model.VersionData{
					Versions: map[string]model.VersionInfo{
						"Default": {
							ExtendedPaths: &model.ExtendedPathsSet{
								URLRewrite: []model.URLRewriteMeta{{RewriteTo: ""}},
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
									{RewriteTo: "", Triggers: []model.RoutingTrigger{{}}},
								},
							},
						},
					},
				},
			},
		},
	}

	errorDetail := "can't be empty"
	missingTargetURLErr := field.Error{
		Type:   field.ErrorTypeRequired,
		Field:  path("proxy", "target_url").String(),
		Detail: errorDetail,
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
			name: "valid ApiDefinition, targeting internal API",
			apiDef: ApiDefinition{
				Spec: APIDefinitionSpec{
					APIDefinitionSpec: model.APIDefinitionSpec{
						Proxy: model.Proxy{TargetURL: "", TargetInternal: &model.TargetInternal{
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
						GraphQL: &model.GraphQLConfig{Enabled: true},
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
				Detail: errorDetail,
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
				Detail: errorDetail,
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

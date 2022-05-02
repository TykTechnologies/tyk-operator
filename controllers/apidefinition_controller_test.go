package controllers

import (
	"encoding/base64"
	"reflect"
	"sort"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

func TestUpdatingLoopingTargets(t *testing.T) {
	t.Run(".spec.proxy.target_internal", func(t *testing.T) {
		t.Skip()
		target := &model.TargetInternal{
			Target: model.Target{
				Name:      "test",
				Namespace: "default",
			},
		}
		a := &tykv1alpha1.ApiDefinition{
			Spec: tykv1alpha1.APIDefinitionSpec{
				APIDefinitionSpec: model.APIDefinitionSpec{
					Proxy: model.Proxy{
						TargetInternal: target,
					},
				},
			},
		}

		got := a.Spec.CollectLoopingTarget()

		if len(got) != 1 {
			t.Fatalf("expected 1 target got %d", len(got))
		}

		if !got[0].Equal(target.Target) {
			t.Fatalf("expected %v got %v", target.Target, got[0])
		}

		// we now check if .spec.proxy.target was updated
		if a.Spec.Proxy.TargetURL != target.String() {
			t.Errorf("expected %q got %q", target.String(), a.Spec.Proxy.TargetURL)
		}
		// make sure we the looping target was set to null
		if a.Spec.Proxy.TargetInternal != nil {
			t.Errorf("expected .spec.proxy.target_internal to be nil ")
		}
	})
	t.Run(".spec.version_data[url_rewrite]", func(t *testing.T) {
		t.Skip()
		target := &model.RewriteToInternal{
			Target: model.Target{
				Name:      "test",
				Namespace: "default",
			},
		}
		a := &tykv1alpha1.ApiDefinition{
			Spec: tykv1alpha1.APIDefinitionSpec{
				APIDefinitionSpec: model.APIDefinitionSpec{
					VersionData: model.VersionData{
						Versions: map[string]model.VersionInfo{
							"Default": {
								ExtendedPaths: &model.ExtendedPathsSet{
									URLRewrite: []model.URLRewriteMeta{
										{
											RewriteToInternal: target,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		got := a.Spec.CollectLoopingTarget()
		if len(got) != 1 {
			t.Fatalf("expected 1 target got %d", len(got))
		}
		if !got[0].Equal(target.Target) {
			t.Fatalf("expected %v got %v", target.Target, got[0])
		}
		meta := a.Spec.VersionData.Versions["Default"].ExtendedPaths.URLRewrite[0]
		if meta.RewriteTo != target.String() {
			t.Errorf("expected %q got %q", target.String(), meta.RewriteTo)
		}
		// make sure we the looping target was set to null
		if meta.RewriteToInternal != nil {
			t.Errorf("expected .spec.proxy.target_internal to be nil ")
		}
	})
	t.Run(".spec.version_data[url_rewrite triggers]", func(t *testing.T) {
		target := &model.RewriteToInternal{
			Target: model.Target{
				Name:      "test",
				Namespace: "default",
			},
		}
		a := &tykv1alpha1.ApiDefinition{
			Spec: tykv1alpha1.APIDefinitionSpec{
				APIDefinitionSpec: model.APIDefinitionSpec{
					VersionData: model.VersionData{
						Versions: map[string]model.VersionInfo{
							"Default": {
								ExtendedPaths: &model.ExtendedPathsSet{
									URLRewrite: []model.URLRewriteMeta{
										{
											Triggers: []model.RoutingTrigger{
												{
													RewriteToInternal: target,
												},
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
		got := a.Spec.CollectLoopingTarget()
		if len(got) != 1 {
			t.Fatalf("expected 1 target got %d", len(got))
		}
		if !got[0].Equal(target.Target) {
			t.Fatalf("expected %v got %v", target.Target, got[0])
		}
		meta := a.Spec.VersionData.Versions["Default"].ExtendedPaths.URLRewrite[0].Triggers[0]
		if meta.RewriteTo != target.String() {
			t.Errorf("expected %q got %q", target.String(), meta.RewriteTo)
		}
		// make sure we the looping target was set to null
		if meta.RewriteToInternal != nil {
			t.Errorf("expected .spec.proxy.target_internal to be nil ")
		}
	})
	t.Run("all", func(t *testing.T) {
		t.Skip()
		target1 := &model.TargetInternal{
			Target: model.Target{
				Name:      "test1",
				Namespace: "default",
			},
		}
		target2 := &model.RewriteToInternal{
			Target: model.Target{
				Name:      "test2",
				Namespace: "default",
			},
		}
		target3 := &model.RewriteToInternal{
			Target: model.Target{
				Name:      "test3",
				Namespace: "default",
			},
		}
		a := &tykv1alpha1.ApiDefinition{
			Spec: tykv1alpha1.APIDefinitionSpec{
				APIDefinitionSpec: model.APIDefinitionSpec{
					Proxy: model.Proxy{
						TargetInternal: target1,
					},
					VersionData: model.VersionData{
						Versions: map[string]model.VersionInfo{
							"Default": {
								ExtendedPaths: &model.ExtendedPathsSet{
									URLRewrite: []model.URLRewriteMeta{
										{
											RewriteToInternal: target2,
											Triggers: []model.RoutingTrigger{
												{
													RewriteToInternal: target3,
												},
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
		got := a.Spec.CollectLoopingTarget()
		if len(got) != 3 {
			t.Fatalf("expected 3 target got %d", len(got))
		}
		expect := []model.Target{
			target1.Target, target2.Target, target3.Target,
		}
		sort.Slice(expect, func(i, j int) bool {
			return expect[i].String() < expect[j].String()
		})
		if !reflect.DeepEqual(got, expect) {
			t.Fatalf("expected %#v got %#v", expect, got)
		}
	})
}

func TestTargetInternal(t *testing.T) {
	target3 := &model.RewriteToInternal{
		Target: model.Target{
			Name:      "test3",
			Namespace: "default",
		},
		Path:  "proxy/$1",
		Query: "a=1&b=2",
	}
	if expert, got := "tyk://ZGVmYXVsdC90ZXN0Mw/proxy/$1?a=1&b=2", target3.String(); expert != got {
		t.Errorf("expected %q got %q", expert, got)
	}
}

func TestEncodeIfNotBase64(t *testing.T) {
	in := "default/httpbin-security-policy"
	s := encodeIfNotBase64(in)

	out, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		t.Fatal(err.Error())
	}

	if string(out) != in {
		t.Fatal("out should be in")
	}

	in = "ZGVmYXVsdC9odHRwYmluLXNlY3VyaXR5LXBvbGljeQ"

	s = encodeIfNotBase64(in)
	if s != in {
		t.Fatalf("expect %s, got %s", in, s)
	}
}

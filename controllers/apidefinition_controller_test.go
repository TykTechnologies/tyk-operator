package controllers

import (
	"reflect"
	"sort"
	"testing"

	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

func TestUpdatingLoopingTargets(t *testing.T) {
	t.Run(".spec.proxy.target_internal", func(t *testing.T) {
		target := &tykv1alpha1.TargetInternal{
			Target: tykv1alpha1.Target{
				Name:      "test",
				Namespace: "default",
			},
		}
		a := &tykv1alpha1.ApiDefinition{
			Spec: tykv1alpha1.APIDefinitionSpec{
				Proxy: tykv1alpha1.Proxy{
					TargeInternal: target,
				},
			},
		}
		got := collectAndUpdateLoopingTargets(a)
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
		if a.Spec.Proxy.TargeInternal != nil {
			t.Errorf("expected .spec.proxy.target_internal to be nil ")
		}
	})
	t.Run(".spec.version_data[url_rewrite]", func(t *testing.T) {
		target := &tykv1alpha1.RewriteToInternal{
			Target: tykv1alpha1.Target{
				Name:      "test",
				Namespace: "default",
			},
		}
		a := &tykv1alpha1.ApiDefinition{
			Spec: tykv1alpha1.APIDefinitionSpec{
				VersionData: tykv1alpha1.VersionData{
					Versions: map[string]tykv1alpha1.VersionInfo{
						"Default": {
							ExtendedPaths: &tykv1alpha1.ExtendedPathsSet{
								URLRewrite: []tykv1alpha1.URLRewriteMeta{
									{
										RewriteToInternal: target,
									},
								},
							},
						},
					},
				},
			},
		}
		got := collectAndUpdateLoopingTargets(a)
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
		target := &tykv1alpha1.RewriteToInternal{
			Target: tykv1alpha1.Target{
				Name:      "test",
				Namespace: "default",
			},
		}
		a := &tykv1alpha1.ApiDefinition{
			Spec: tykv1alpha1.APIDefinitionSpec{
				VersionData: tykv1alpha1.VersionData{
					Versions: map[string]tykv1alpha1.VersionInfo{
						"Default": {
							ExtendedPaths: &tykv1alpha1.ExtendedPathsSet{
								URLRewrite: []tykv1alpha1.URLRewriteMeta{
									{
										Triggers: []tykv1alpha1.RoutingTrigger{
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
		}
		got := collectAndUpdateLoopingTargets(a)
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
		target1 := &tykv1alpha1.TargetInternal{
			Target: tykv1alpha1.Target{
				Name:      "test1",
				Namespace: "default",
			},
		}
		target2 := &tykv1alpha1.RewriteToInternal{
			Target: tykv1alpha1.Target{
				Name:      "test2",
				Namespace: "default",
			},
		}
		target3 := &tykv1alpha1.RewriteToInternal{
			Target: tykv1alpha1.Target{
				Name:      "test3",
				Namespace: "default",
			},
		}
		a := &tykv1alpha1.ApiDefinition{
			Spec: tykv1alpha1.APIDefinitionSpec{
				Proxy: tykv1alpha1.Proxy{
					TargeInternal: target1,
				},
				VersionData: tykv1alpha1.VersionData{
					Versions: map[string]tykv1alpha1.VersionInfo{
						"Default": {
							ExtendedPaths: &tykv1alpha1.ExtendedPathsSet{
								URLRewrite: []tykv1alpha1.URLRewriteMeta{
									{
										RewriteToInternal: target2,
										Triggers: []tykv1alpha1.RoutingTrigger{
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
		}
		got := collectAndUpdateLoopingTargets(a)
		if len(got) != 3 {
			t.Fatalf("expected 3 target got %d", len(got))
		}
		expect := []tykv1alpha1.Target{
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

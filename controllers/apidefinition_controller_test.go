package controllers

import (
	"context"
	"encoding/base64"
	"reflect"
	"sort"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	tykv1alpha1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"

	"github.com/matryer/is"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

func TestProcessSubGraphExecution(t *testing.T) {
	const (
		subgraphName = "test-subgraph"
		testSDL      = "sdl"
		testSchema   = "schema"

		newSubgraphName = "test-subgraph-new"
		newTestSDL      = "new-sdl"
		newTestSchema   = "new-schema"

		testNs = "test-ns"
	)

	subGraph := &tykv1alpha1.SubGraph{
		ObjectMeta: v1.ObjectMeta{
			Name:      subgraphName,
			Namespace: testNs,
		},
		Spec: tykv1alpha1.SubGraphSpec{
			SubGraphSpec: model.SubGraphSpec{
				SDL:    testSDL,
				Schema: testSchema,
			},
		},
	}
	newSubGraph := &tykv1alpha1.SubGraph{
		ObjectMeta: v1.ObjectMeta{
			Name:      newSubgraphName,
			Namespace: testNs,
		},
		Spec: tykv1alpha1.SubGraphSpec{
			SubGraphSpec: model.SubGraphSpec{
				SDL:    newTestSDL,
				Schema: newTestSchema,
			},
		},
	}

	apiDef := tykv1alpha1.ApiDefinition{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-name",
			Namespace: testNs,
		},
		Spec: tykv1alpha1.APIDefinitionSpec{
			APIDefinitionSpec: model.APIDefinitionSpec{
				APIID: encodeNS(types.NamespacedName{Name: "test-name", Namespace: testNs}.String()),
				GraphQL: &model.GraphQLConfig{
					GraphRef: subgraphName,
				},
			},
		},
	}

	apiDefMalformed := apiDef
	apiDefMalformed.Spec.GraphQL = nil

	apiDefLinkedMultiple := &tykv1alpha1.ApiDefinition{}
	apiDef.DeepCopyInto(apiDefLinkedMultiple)
	apiDefLinkedMultiple.ObjectMeta.Name = "test-name-multiple-link"
	apiDefLinkedMultiple.Spec.APIID = encodeNS(client.ObjectKeyFromObject(apiDefLinkedMultiple).String())

	objects := []runtime.Object{&apiDef, apiDefLinkedMultiple, subGraph, newSubGraph}
	eval := is.New(t)

	cl, err := NewFakeClient(objects)
	eval.NoErr(err)

	testCases := []struct {
		testName       string
		expectedSDL    string
		expectedSchema string
		apiDef         *tykv1alpha1.ApiDefinition
		subGraph       *tykv1alpha1.SubGraph
		mutateFn       func(definition *tykv1alpha1.ApiDefinition)
		expectedErr    error
	}{
		{
			testName: "processing malformed ApiDefinition with nil GraphQLConfig field",
			apiDef:   &apiDefMalformed,
		},
		{
			testName: "processing valid ApiDefinition must update it's GraphQLConfig based on SubGraph reference",
			apiDef:   &apiDef,
			subGraph: subGraph,
		},
		{
			testName:    "processing ApiDefinition referencing already linked SubGraph CR must fail",
			apiDef:      apiDefLinkedMultiple,
			subGraph:    subGraph,
			expectedErr: ErrMultipleLinkSubGraph,
		},
		{
			testName: "update ApiDefinition GraphRef to another SubGraph CR",
			apiDef:   &apiDef,
			subGraph: newSubGraph,
			mutateFn: func(definition *tykv1alpha1.ApiDefinition) {
				definition.Spec.GraphQL.GraphRef = newSubgraphName
			},
		},
		{
			testName: "remove GraphRef from ApiDefinition CR",
			apiDef:   &apiDef,
			// subGraph: newSubGraph,
			subGraph: &tykv1alpha1.SubGraph{},
			mutateFn: func(definition *tykv1alpha1.ApiDefinition) {
				definition.Spec.GraphQL.GraphRef = ""
				definition.Spec.APIID = ""
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			r := ApiDefinitionReconciler{
				Client: cl,
				Scheme: scheme.Scheme,
				Log:    log.NullLogger{},
			}

			api := &tykv1alpha1.ApiDefinition{}
			err = r.Client.Get(context.Background(), client.ObjectKeyFromObject(tc.apiDef), api)
			eval.NoErr(err)

			if tc.mutateFn != nil {
				tc.mutateFn(api)
			}

			err = r.processSubGraphExec(context.Background(), api)
			if _, ok := err.(*k8sErrors.StatusError); ok {
				eval.Equal(tc.expectedErr.Error(), string(k8sErrors.ReasonForError(err)))
			} else {
				eval.Equal(tc.expectedErr, err)
			}

			if tc.expectedErr == nil && tc.apiDef.Spec.GraphQL != nil {
				eval.Equal(tc.subGraph.Spec.Schema, api.Spec.GraphQL.Schema)
				eval.Equal(tc.subGraph.Spec.SDL, api.Spec.GraphQL.Subgraph.SDL)

				reconciledApiDef := &tykv1alpha1.ApiDefinition{}
				err = r.Client.Get(context.Background(), client.ObjectKeyFromObject(tc.apiDef), reconciledApiDef)
				eval.NoErr(err)
				eval.Equal(reconciledApiDef.Status.LinkedSubgraphName, tc.subGraph.Name)

				reconciledSubGraph := &tykv1alpha1.SubGraph{}
				err = r.Client.Get(context.Background(), client.ObjectKeyFromObject(tc.subGraph), reconciledSubGraph)
				eval.NoErr(client.IgnoreNotFound(err))
				eval.Equal(reconciledSubGraph.Status.LinkedApiDefID, api.Spec.APIID)
			}
		})
	}
}

func TestDecodeID(t *testing.T) {
	type args struct {
		encodedID string
	}
	tests := []struct {
		name         string
		encodedID    string
		expectedNs   string
		expectedName string
	}{
		{
			name:         "decoding empty ID",
			encodedID:    "",
			expectedNs:   "",
			expectedName: "",
		},
		{
			name:         "decoding default/httpbin",
			encodedID:    "ZGVmYXVsdC9odHRwYmlu",
			expectedNs:   "default",
			expectedName: "httpbin",
		},
		{
			name:         "decoding corrupted input, /httpbin",
			encodedID:    "L2h0dHBiaW4=",
			expectedNs:   "",
			expectedName: "",
		},
		{
			name:         "decoding corrupted input, default/",
			encodedID:    "ZGVmYXVsdC8=",
			expectedNs:   "",
			expectedName: "",
		},
		{
			name:         "decoding corrupted input, /",
			encodedID:    "Lw==",
			expectedNs:   "",
			expectedName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNamespace, gotName := decodeID(tt.encodedID)
			if gotNamespace != tt.expectedNs {
				t.Errorf("decodeID() gotNamespace = %v, want %v", gotNamespace, tt.expectedNs)
			}

			if gotName != tt.expectedName {
				t.Errorf("decodeID() gotName = %v, want %v", gotName, tt.expectedName)
			}
		})
	}
}

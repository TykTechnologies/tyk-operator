package controllers

import (
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

func TestDecodeID(t *testing.T) {
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

func TestChanged(t *testing.T) {
	type args struct {
		latestHash string
		i1         interface{}
	}

	nilHash, emptyHash := calculateHashes(nil, v1alpha1.ApiDefinition{})

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			args: args{
				latestHash: nilHash,
				i1:         nil,
			},
			want: false,
		},
		{
			args: args{
				latestHash: emptyHash,
				i1:         v1alpha1.ApiDefinition{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := changed(tt.args.latestHash, tt.args.i1); got != tt.want {
				t.Errorf("changed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSameApiDefinition(t *testing.T) {
	type args struct {
		crd1 *v1alpha1.ApiDefinition
		crd2 *v1alpha1.ApiDefinition
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Checking two nil ApiDefinitionSpec objects",
			args: args{
				crd1: nil,
				crd2: nil,
			},
			want: false,
		},
		{
			name: "Checking two empty ApiDefinitionSpec objects",
			args: args{
				crd1: &v1alpha1.ApiDefinition{},
				crd2: &v1alpha1.ApiDefinition{},
			},
			want: false,
		},
		{
			name: "Checking two non-nil same ApiDefinitionSpec objects",
			args: args{
				crd1: &v1alpha1.ApiDefinition{
					Spec: v1alpha1.APIDefinitionSpec{
						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
					},
				},
				crd2: &v1alpha1.ApiDefinition{
					Spec: v1alpha1.APIDefinitionSpec{
						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
					},
				},
			},
			want: false,
		},
		{
			name: "Checking two non-nil different ApiDefinitionSpec objects",
			args: args{
				crd1: &v1alpha1.ApiDefinition{
					Spec: v1alpha1.APIDefinitionSpec{
						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API2"},
					},
				},
				crd2: &v1alpha1.ApiDefinition{
					Spec: v1alpha1.APIDefinitionSpec{
						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
					},
				},
			},
			want: true,
		},
		{
			name: "Checking empty CRD and non-empty ApiDefinitionSpec",
			args: args{
				crd1: &v1alpha1.ApiDefinition{},
				crd2: &v1alpha1.ApiDefinition{
					Spec: v1alpha1.APIDefinitionSpec{
						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
					},
				},
			},
			want: true,
		},
		{
			name: "Checking non-empty CRD and empty ApiDefinitionSpec",
			args: args{
				crd1: &v1alpha1.ApiDefinition{
					Spec: v1alpha1.APIDefinitionSpec{
						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
					},
				},
				crd2: &v1alpha1.ApiDefinition{
					Spec: v1alpha1.APIDefinitionSpec{
						APIDefinitionSpec: model.APIDefinitionSpec{},
					},
				},
			},
			want: true,
		},
		{
			name: "Checking nil CRD and non-empty ApiDefinitionSpec",
			args: args{
				crd1: nil,
				crd2: &v1alpha1.ApiDefinition{
					Spec: v1alpha1.APIDefinitionSpec{
						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
					},
				},
			},
			want: true,
		},
		{
			name: "Checking non-empty CRD and nil ApiDefinitionSpec",
			args: args{
				crd1: &v1alpha1.ApiDefinition{
					Spec: v1alpha1.APIDefinitionSpec{
						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
					},
				},
				crd2: nil,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1, _ := calculateHashes(tt.args.crd1, nil)

			got := changed(h1, tt.args.crd2)
			if got != tt.want {
				t.Errorf("changed() = %v, want %v", got, tt.want)
			}
		})
	}
}

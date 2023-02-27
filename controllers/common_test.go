package controllers

import (
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/model"
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

func TestIsSameApiDefinition(t *testing.T) {
	type args struct {
		apiDef1 *model.APIDefinitionSpec
		apiDef2 *model.APIDefinitionSpec
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Checking two nil ApiDefinitionSpec objects ",
			args: args{
				apiDef1: &model.APIDefinitionSpec{},
				apiDef2: &model.APIDefinitionSpec{},
			},
			want: true,
		},
		{
			name: "Checking two non-nil same ApiDefinitionSpec objects ",
			args: args{
				apiDef1: &model.APIDefinitionSpec{Name: "My API"},
				apiDef2: &model.APIDefinitionSpec{Name: "My API"},
			},
			want: true,
		},
		{
			name: "Checking two non-nil different ApiDefinitionSpec objects ",
			args: args{
				apiDef1: &model.APIDefinitionSpec{Name: "My API2"},
				apiDef2: &model.APIDefinitionSpec{Name: "My API"},
			},
			want: false,
		},
		{
			name: "Checking one nil and one non-nil ApiDefinitionSpec objects ",
			args: args{
				apiDef1: &model.APIDefinitionSpec{},
				apiDef2: &model.APIDefinitionSpec{Name: "My API"},
			},
			want: false,
		},
		{
			name: "Checking whether ID field is ignored",
			args: args{
				apiDef1: &model.APIDefinitionSpec{ID: "sample"},
				apiDef2: &model.APIDefinitionSpec{ID: "different"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSameApiDefinition(tt.args.apiDef1, tt.args.apiDef2); got != tt.want {
				t.Errorf("isSameApiDefinition() = %v, want %v", got, tt.want)
			}
		})
	}
}

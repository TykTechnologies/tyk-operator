package controllers

import (
	"testing"
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

//func TestIsSameApiDefinition(t *testing.T) {
//	type args struct {
//		crdApi *v1alpha1.ApiDefinition
//		tykApi *model.APIDefinitionSpec
//	}
//	tests := []struct {
//		name string
//		args args
//		want bool
//	}{
//		{
//			name: "Checking two nil ApiDefinitionSpec objects",
//			args: args{
//				crdApi: nil,
//				tykApi: nil,
//			},
//			want: true,
//		},
//		{
//			name: "Checking two empty ApiDefinitionSpec objects",
//			args: args{
//				crdApi: &v1alpha1.ApiDefinition{},
//				tykApi: &model.APIDefinitionSpec{},
//			},
//			want: true,
//		},
//		{
//			name: "Checking two non-nil same ApiDefinitionSpec objects",
//			args: args{
//				crdApi: &v1alpha1.ApiDefinition{
//					Spec: v1alpha1.APIDefinitionSpec{
//						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
//					},
//				},
//				tykApi: &model.APIDefinitionSpec{Name: "My API"},
//			},
//			want: true,
//		},
//		{
//			name: "Checking two non-nil different ApiDefinitionSpec objects",
//			args: args{
//				crdApi: &v1alpha1.ApiDefinition{
//					Spec: v1alpha1.APIDefinitionSpec{
//						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API2"},
//					},
//				},
//				tykApi: &model.APIDefinitionSpec{Name: "My API"},
//			},
//			want: false,
//		},
//		{
//			name: "Checking empty CRD and non-empty ApiDefinitionSpec",
//			args: args{
//				crdApi: &v1alpha1.ApiDefinition{},
//				tykApi: &model.APIDefinitionSpec{Name: "My API"},
//			},
//			want: false,
//		},
//		{
//			name: "Checking non-empty CRD and empty ApiDefinitionSpec",
//			args: args{
//				crdApi: &v1alpha1.ApiDefinition{
//					Spec: v1alpha1.APIDefinitionSpec{
//						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
//					},
//				},
//				tykApi: &model.APIDefinitionSpec{},
//			},
//			want: false,
//		},
//		{
//			name: "Checking nil CRD and non-empty ApiDefinitionSpec",
//			args: args{
//				crdApi: nil,
//				tykApi: &model.APIDefinitionSpec{Name: "My API"},
//			},
//			want: false,
//		},
//		{
//			name: "Checking non-empty CRD and nil ApiDefinitionSpec",
//			args: args{
//				crdApi: &v1alpha1.ApiDefinition{
//					Spec: v1alpha1.APIDefinitionSpec{
//						APIDefinitionSpec: model.APIDefinitionSpec{Name: "My API"},
//					},
//				},
//				tykApi: nil,
//			},
//			want: false,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// before comparing whether resources are same, update CRD status with correct hashes.
//			if tt.args.crdApi != nil {
//				crdHash, _ := calculateHashes(tt.args.crdApi.Spec.APIDefinitionSpec, nil)
//				tt.args.crdApi.Status.LatestCRDHash = crdHash
//			}
//
//			if tt.args.tykApi != nil {
//				tykHash, _ := calculateHashes(tt.args.tykApi, nil)
//				tt.args.crdApi.Status.LatestTykHash = tykHash
//			}
//
//			if got := isSameApiDefinition(tt.args.crdApi, tt.args.tykApi); got != tt.want {
//				t.Errorf("isSameApiDefinition() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

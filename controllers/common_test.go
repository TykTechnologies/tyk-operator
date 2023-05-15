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

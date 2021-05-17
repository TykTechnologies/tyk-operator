package controllers

import (
	"encoding/base64"

	tykv1 "github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

// Helper function to check string exists in a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// Helper function to remove string from slice of string
func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// addString returns a string slice with s in it. If s is in slice then slice
// will be returned
func addString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			return slice
		}
	}
	return append(slice, s)
}

func addTarget(slice []tykv1.Target, s tykv1.Target) (result []tykv1.Target) {
	for _, item := range slice {
		if item == s {
			return slice
		}
	}
	return append(slice, s)
}

func removeTarget(slice []tykv1.Target, s tykv1.Target) (result []tykv1.Target) {
	for _, item := range slice {
		if item.Equal(s) {
			continue
		}
		result = append(result, item)
	}
	return
}

func encodeNS(decoded string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(decoded))
}

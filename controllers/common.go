package controllers

import (
	"context"
	"encoding/base64"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/logr"
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

func addTarget(slice []model.Target, s model.Target) (result []model.Target) {
	for _, item := range slice {
		if item == s {
			return slice
		}
	}
	return append(slice, s)
}

func removeTarget(slice []model.Target, s model.Target) (result []model.Target) {
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

func httpContext(ctx context.Context, e environmet.Env, log logr.Logger) context.Context {
	return client.SetContext(ctx, client.Context{
		Env: e,
		Log: log,
	})
}

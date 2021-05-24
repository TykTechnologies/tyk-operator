package controllers

import (
	"context"
	"encoding/base64"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// GetContext returns a OperatorContext context resource from k8s cluster with
// target. When Spec.FromSecret is provided this reads the secret and loads the
// environment from it. Values set in .Spec.Env takes precedence over the values
// from secret
func GetContext(
	ctx context.Context,
	client client.Client,
	target *model.Target,
) (*v1alpha1.OperatorContext, error) {
	var o v1alpha1.OperatorContext
	err := client.Get(ctx, target.NS(), &o)
	if err != nil {
		return nil, err
	}
	if o.Spec.FromSecret != nil {
		var secret v1.Secret
		if err := client.Get(ctx, o.Spec.FromSecret.NS(), &secret); err != nil {
			return nil, err
		}
	}
	if o.Spec.Env == nil {
		o.Spec.Env = &v1alpha1.Environment{}
	}
	return &o, nil
}

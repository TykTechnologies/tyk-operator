package controllers

import (
	"context"
	"encoding/base64"
	"strconv"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
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

func httpContext(
	ctx context.Context,
	rClient runtimeClient.Client,
	e environmet.Env,
	object runtimeClient.Object,
	log logr.Logger,
) (environmet.Env, context.Context, error) {
	get := func(c *model.Target) error {
		if c == nil {
			// To handle the case where operator context was used previously
			// but was removed in update operation
			if err := updateOperatorContextStatus(ctx, rClient, e, object, log, c); err != nil {
				log.Error(err, "Failed to update status of operator contexts")
			}

			return nil
		}

		// If namespace is not specified in contextDef, use default namespace
		if c.Namespace == "" {
			log.Info("Context namespace is not specified, using default")
			c.Namespace = "default"
		}

		log.Info("Detected context for resource")

		env, err := GetContext(
			ctx, object.GetNamespace(), rClient, c, log,
		)
		if err != nil {
			log.Error(err, "Failed to get context", "contextRef", c.String())

			return err
		}

		log.Info("Successful acquired context", "contextRef", c.String())

		e.Environment = *env.Spec.Env

		if err := updateOperatorContextStatus(ctx, rClient, e, object, log, c); err != nil {
			log.Error(err, "Failed to update status of operator contexts")
		}

		return nil
	}

	var err error

	switch o := object.(type) {
	case *v1alpha1.ApiDefinition:
		err = get(o.Spec.Context)
	case *v1alpha1.SecurityPolicy:
		err = get(o.Spec.Context)
	case *v1alpha1.PortalAPICatalogue:
		err = get(o.Spec.Context)
	case *v1alpha1.APIDescription:
		err = get(o.Spec.Context)
	case *v1alpha1.PortalConfig:
		err = get(o.Spec.Context)
	}

	if err != nil {
		return environmet.Env{}, nil, err
	}

	return e, client.SetContext(ctx, client.Context{
		Env: e,
		Log: log,
	}), nil
}

func updateOperatorContextStatus(
	ctx context.Context,
	rClient runtimeClient.Client,
	e environmet.Env,
	object runtimeClient.Object,
	log logr.Logger,
	ctxRef *model.Target,
) error {
	target := model.Target{
		Name:      object.GetName(),
		Namespace: object.GetNamespace(),
	}

	// Remove link from other operator context, if any,
	// as we do not know if object was referencing to different context previously
	var opCtxList v1alpha1.OperatorContextList

	if err := rClient.List(ctx, &opCtxList); err != nil {
		return err
	}

	switch object.(type) {
	case *v1alpha1.ApiDefinition:
		for _, opctx := range opCtxList.Items {
			// do not remove link if apidef is still refering to same context and is not marked for deletion
			if ctxRef != nil && opctx.Name == ctxRef.Name && opctx.Namespace == ctxRef.Namespace &&
				object.GetDeletionTimestamp().IsZero() {
				continue
			}

			opctx.Status.RemoveLinkedAPIDefinition(target)

			err := rClient.Status().Update(ctx, &opctx)
			if err != nil {
				log.Error(err, "Failed to remove link of APIDefintion from operator context", "operatorContext", opctx.Name, "apidefinition", target.Name)
			}
		}
	case *v1alpha1.SecurityPolicy:
		for _, opctx := range opCtxList.Items {
			// do not remove link if apidef is still refering to context and is not marked for deletion
			if ctxRef != nil && opctx.Name == ctxRef.Name && opctx.Namespace == ctxRef.Namespace &&
				object.GetDeletionTimestamp().IsZero() {
				continue
			}

			opctx.Status.RemoveLinkedSecurityPolicies(target)

			err := rClient.Status().Update(ctx, &opctx)
			if err != nil {
				return err
			}
		}
	case *v1alpha1.PortalAPICatalogue:
		for _, opctx := range opCtxList.Items {
			// do not remove link if apidef is still refering to context and is not marked for deletion
			if ctxRef != nil && opctx.Name == ctxRef.Name && opctx.Namespace == ctxRef.Namespace &&
				object.GetDeletionTimestamp().IsZero() {
				continue
			}

			opctx.Status.RemoveLinkedPortalAPICatalogues(target)

			err := rClient.Status().Update(ctx, &opctx)
			if err != nil {
				return err
			}
		}
	case *v1alpha1.APIDescription:
		for _, opctx := range opCtxList.Items {
			// do not remove link if apidef is still refering to context and is not marked for deletion
			if ctxRef != nil && opctx.Name == ctxRef.Name && opctx.Namespace == ctxRef.Namespace &&
				object.GetDeletionTimestamp().IsZero() {
				continue
			}

			opctx.Status.RemoveLinkedApiDescriptions(target)

			err := rClient.Status().Update(ctx, &opctx)
			if err != nil {
				return err
			}
		}
	case *v1alpha1.PortalConfig:
		for _, opctx := range opCtxList.Items {
			// do not remove link if apidef is still refering to context and is not marked for deletion
			if ctxRef != nil && opctx.Name == ctxRef.Name && opctx.Namespace == ctxRef.Namespace &&
				object.GetDeletionTimestamp().IsZero() {
				continue
			}

			opctx.Status.RemoveLinkedPortalConfig(target)

			err := rClient.Status().Update(ctx, &opctx)
			if err != nil {
				return err
			}
		}
	}

	// Add reference to the refered operator context
	// only if object is not marked for deletion
	if object.GetDeletionTimestamp().IsZero() && ctxRef != nil {
		// add reference to operator context
		log.Info("Adding link to apiContext", "key", ctxRef.String())

		var operatorContext v1alpha1.OperatorContext

		key := types.NamespacedName{Name: ctxRef.Name, Namespace: ctxRef.Namespace}

		if err := rClient.Get(ctx, key, &operatorContext); err != nil {
			log.Error(err, "failed to get operator context")
			return err
		}

		switch object.(type) {
		case *v1alpha1.ApiDefinition:
			operatorContext.Status.AddLinkedAPIDefinition(target)
		case *v1alpha1.SecurityPolicy:
			operatorContext.Status.AddLinkedSecurityPolicies(target)
		case *v1alpha1.PortalAPICatalogue:
			operatorContext.Status.AddLinkedPortalAPICatalogues(target)
		case *v1alpha1.APIDescription:
			operatorContext.Status.AddLinkedApiDescriptions(target)
		case *v1alpha1.PortalConfig:
			operatorContext.Status.AddLinkedPortalConfig(target)
		}

		return rClient.Status().Update(ctx, &operatorContext)
	}

	return nil
}

// +kubebuilder:rbac:groups=tyk.tyk.io,resources=operatorcontexts,verbs=get;list;watch;create;update;patch;delete

// GetContext returns a OperatorContext resource from k8s cluster with
// namespace/name derived from target. When Spec.FromSecret is provided this
// reads the secret and loads the environment from it. Values set in .Spec.Env
// takes precedence over the values from secret
func GetContext(
	ctx context.Context,
	ns string,
	client runtimeClient.Client,
	target *model.Target,
	log logr.Logger,
) (*v1alpha1.OperatorContext, error) {
	newTarget := target.NS(ns)

	log.Info("Getting context", "contextRef", newTarget.String())

	var o v1alpha1.OperatorContext

	err := client.Get(ctx, newTarget, &o)
	if err != nil {
		return nil, err
	}

	if o.Spec.Env == nil {
		o.Spec.Env = &v1alpha1.Environment{}
	}

	if o.Spec.FromSecret != nil {
		var secret v1.Secret

		if err := client.Get(ctx, o.Spec.FromSecret.NS(o.Namespace), &secret); err != nil {
			return nil, err
		}

		value := func(key string, fn func(string) error) error {
			if v, ok := secret.Data[key]; ok {
				return fn(string(v))
			}

			return nil
		}
		// we are setting all values that are not set on env but present in secret.
		e := o.Spec.Env
		if e.Mode == "" {
			err := value(v1alpha1.TykMode, func(s string) error {
				e.Mode = v1alpha1.OperatorContextMode(s)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}

		if !e.InsecureSkipVerify {
			err = value(v1alpha1.SkipVerify, func(s string) (err error) {
				e.InsecureSkipVerify, err = strconv.ParseBool(s)
				return
			})
			if err != nil {
				return nil, err
			}
		}

		if e.URL == "" {
			err := value(v1alpha1.TykURL, func(s string) (err error) {
				e.URL = s
				return
			})
			if err != nil {
				return nil, err
			}
		}

		if e.Auth == "" {
			err := value(v1alpha1.TykAuth, func(s string) (err error) {
				e.Auth = s
				return
			})
			if err != nil {
				return nil, err
			}
		}

		if e.Org == "" {
			err := value(v1alpha1.TykORG, func(s string) (err error) {
				e.Org = s
				return
			})
			if err != nil {
				return nil, err
			}
		}

		if e.Ingress.HTTPPort == 0 {
			err = value(v1alpha1.IngressHTTPPort, func(s string) (err error) {
				e.Ingress.HTTPPort, err = strconv.Atoi(s)
				return
			})
			if err != nil {
				return nil, err
			}
		}

		if e.Ingress.HTTPSPort == 0 {
			err = value(v1alpha1.IngressTLSPort, func(s string) (err error) {
				e.Ingress.HTTPSPort, err = strconv.Atoi(s)
				return
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return &o, nil
}

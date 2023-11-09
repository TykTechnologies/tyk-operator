package environment

import (
	"os"
	"strconv"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

// Env holds values needed to talk to the gateway or the dashboard API
type Env struct {
	v1alpha1.Environment
	model.EnvVariable
	Namespace    string
	IngressClass string
	TykVersion   string
}

func (e Env) Merge(n Env) Env {
	if n.Namespace != "" {
		e.Namespace = n.Namespace
	}

	if n.Mode != "" {
		e.Mode = n.Mode
	}

	if n.URL != "" {
		e.URL = n.URL
	}

	if n.Auth != "" {
		e.Auth = n.Auth
	}

	if n.Org != "" {
		e.Org = n.Org
	}

	if n.IngressClass != "" {
		e.IngressClass = n.IngressClass
	}

	if n.Ingress.HTTPSPort != 0 {
		e.Ingress.HTTPSPort = n.Ingress.HTTPSPort
	}

	if n.Ingress.HTTPPort != 0 {
		e.Ingress.HTTPPort = n.Ingress.HTTPPort
	}

	if n.Name != "" {
		e.Name = n.Name
	}

	if n.ValueFrom.SecretKeyRef.Name != "" {
		e.ValueFrom.SecretKeyRef.Name = n.ValueFrom.SecretKeyRef.Name
	}

	if n.ValueFrom.SecretKeyRef.Key != "" {
		e.ValueFrom.SecretKeyRef.Key = n.ValueFrom.SecretKeyRef.Key
	}

	if n.ValueFrom.ConfigMapKeyRef.Name != "" {
		e.ValueFrom.ConfigMapKeyRef.Name = n.ValueFrom.ConfigMapKeyRef.Name
	}

	if n.ValueFrom.ConfigMapKeyRef.Key != "" {
		e.ValueFrom.ConfigMapKeyRef.Key = n.ValueFrom.ConfigMapKeyRef.Key
	}

	if n.UserOwners != nil {
		e.UserOwners = append(e.UserOwners, n.UserOwners...)
	}

	if n.UserGroupOwners != nil {
		e.UserGroupOwners = append(e.UserGroupOwners, n.UserGroupOwners...)
	}

	return e
}

// Parse loads env vars into e.
func (e *Env) Parse() {
	e.Namespace = strings.TrimSpace(os.Getenv(v1alpha1.WatchNamespace))
	e.Mode = v1alpha1.OperatorContextMode(os.Getenv(v1alpha1.TykMode))
	e.URL = strings.TrimSpace(os.Getenv(v1alpha1.TykURL))
	e.Auth = strings.TrimSpace(os.Getenv(v1alpha1.TykAuth))
	e.Org = strings.TrimSpace(os.Getenv(v1alpha1.TykORG))
	e.InsecureSkipVerify, _ = strconv.ParseBool(os.Getenv(v1alpha1.SkipVerify))
	e.Ingress.HTTPSPort, _ = strconv.Atoi(os.Getenv(v1alpha1.IngressTLSPort))
	e.Ingress.HTTPPort, _ = strconv.Atoi(os.Getenv(v1alpha1.IngressHTTPPort))
	e.IngressClass = os.Getenv(v1alpha1.IngressClass)
	e.Name = strings.TrimSpace(os.Getenv(model.EnvName))
	e.ValueFrom.SecretKeyRef.Name = strings.TrimSpace(os.Getenv(model.SecretRefName))
	e.ValueFrom.SecretKeyRef.Key = strings.TrimSpace(os.Getenv(model.SecretRefKey))
	e.ValueFrom.ConfigMapKeyRef.Name = strings.TrimSpace(os.Getenv(model.ConfigMapRefName))
	e.ValueFrom.ConfigMapKeyRef.Key = strings.TrimSpace(os.Getenv(model.ConfigMapRefKey))

	for _, user := range strings.Split(os.Getenv(v1alpha1.TykUserOwners), ",") {
		if o := strings.TrimSpace(user); o != "" {
			e.UserOwners = append(e.UserOwners, o)
		}
	}

	for _, group := range strings.Split(os.Getenv(v1alpha1.TykUserGroupOwners), ",") {
		if o := strings.TrimSpace(group); o != "" {
			e.UserGroupOwners = append(e.UserGroupOwners, o)
		}
	}

	if e.Ingress.HTTPSPort == 0 {
		e.Ingress.HTTPSPort = 8443
	}
}

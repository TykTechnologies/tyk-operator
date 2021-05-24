package environmet

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
)

// Env holds values needed to talk to the gateway or the dashboard API
type Env struct {
	v1alpha1.Environment
	Namespace    string
	IngressClass string
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
	return e
}

// Parse loads env vars into e and validates them, returning an error if validation fails.
func (e *Env) Parse() error {
	e.Namespace = strings.TrimSpace(os.Getenv(v1alpha1.WatchNamespace))
	e.Mode = v1alpha1.OperatorContextMode(os.Getenv(v1alpha1.TykMode))
	e.URL = strings.TrimSpace(os.Getenv(v1alpha1.TykURL))
	e.Auth = strings.TrimSpace(os.Getenv(v1alpha1.TykAuth))
	e.Org = strings.TrimSpace(os.Getenv(v1alpha1.TykORG))
	e.InsecureSkipVerify, _ = strconv.ParseBool(os.Getenv(v1alpha1.SkipVerify))
	e.Ingress.HTTPSPort, _ = strconv.Atoi(os.Getenv(v1alpha1.IngressTLSPort))
	e.Ingress.HTTPPort, _ = strconv.Atoi(os.Getenv(v1alpha1.IngressHTTPPort))
	e.IngressClass = os.Getenv(v1alpha1.IngressClass)
	if e.Ingress.HTTPSPort == 0 {
		e.Ingress.HTTPSPort = 8443
	}
	// verify
	sample := []struct {
		env, value string
	}{
		{v1alpha1.TykMode, string(e.Mode)},
		{v1alpha1.TykURL, e.URL},
		{v1alpha1.TykAuth, e.Auth},
		{v1alpha1.TykORG, e.Org},
	}
	var ls []string
	for _, v := range sample {
		if v.value == "" {
			ls = append(ls, v.env)
		}
	}
	if len(ls) > 0 {
		return fmt.Errorf("environment vars %v are missing", ls)
	}
	switch e.Mode {
	case "oss", "ce", "pro":
		return nil
	default:
		return fmt.Errorf("unknown TYK_MODE value %q", e.Mode)
	}
}

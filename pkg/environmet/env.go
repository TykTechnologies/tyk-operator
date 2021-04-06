package environmet

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (

	// WatchNamespace is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	WatchNamespace = "WATCH_NAMESPACE"

	// TykMode defines what environment the operator is running. The values are ce
	// for community edition and pro for pro version
	TykMode = "TYK_MODE"

	// TykURL holds the url to either tyk gateway or tyk dashboard
	TykURL = "TYK_URL"

	// TykAuth holds the authorization token used to make api calls to the
	// gateway/dashboard
	TykAuth = "TYK_AUTH"

	// TykORG holds the org id which perform api tasks with
	TykORG = "TYK_ORG"

	// SkipVerify the client will skip tls verification if this is true
	SkipVerify = "TYK_TLS_INSECURE_SKIP_VERIFY"

	// IngressClass overides the default class to watch for ingress
	IngressClass = "WATCH_INGRESS_CLASS"
)

// Env holds values needed to talk to the gateway or the dashboard API
type Env struct {
	Namespace          string
	Mode               string
	InsecureSkipVerify bool
	URL                string
	Auth               string
	Org                string
	IngressClass       string
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
	return e
}

// Parse loads env vars into e and validates them, returning an error if validation fails.
func (e *Env) Parse() error {
	e.Namespace = strings.TrimSpace(os.Getenv(WatchNamespace))
	e.Mode = strings.TrimSpace(os.Getenv(TykMode))
	e.URL = strings.TrimSpace(os.Getenv(TykURL))
	e.Auth = strings.TrimSpace(os.Getenv(TykAuth))
	e.Org = strings.TrimSpace(os.Getenv(TykORG))
	e.InsecureSkipVerify, _ = strconv.ParseBool(os.Getenv(SkipVerify))
	e.IngressClass = os.Getenv(IngressClass)
	// verify
	sample := []struct {
		env, value string
	}{
		{TykMode, e.Mode},
		{TykURL, e.URL},
		{TykAuth, e.Auth},
		{TykORG, e.Org},
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
	case "ce", "pro":
		return nil
	default:
		return fmt.Errorf("unknown TYK_MODE value %q", e.Mode)
	}
}

// JoinURL returns addition of  parts to the base e.URL
func (e *Env) JoinURL(parts ...string) string {
	return joinURL(append([]string{e.URL}, parts...)...)
}

func joinURL(parts ...string) string {
	l := len(parts)
	if l == 1 {
		return parts[0]
	}
	ps := make([]string, l)
	for i, part := range parts {
		if i == 0 {
			ps[i] = strings.TrimRight(part, "/")
		} else {
			ps[i] = strings.TrimLeft(part, "/")
		}
	}
	return strings.Join(ps, "/")
}

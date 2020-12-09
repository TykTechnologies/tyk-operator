package universal_client

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
// which specifies the Namespace to watch.
// An empty value means the operator is running with cluster scope.
const WatchNamespaceEnvVar = "WATCH_NAMESPACE"

type UniversalClient interface {
	HotReload() error
	Api() UniversalApi
	SecurityPolicy() UniversalSecurityPolicy
	Webhook() UniversalWebhook
	Certificate() UniversalCertificate
	Organization() UniversalOrganization
}

// Env holds values needed to talk to the gateway or the dashboard API
type Env struct {
	Namespace          string
	Mode               string
	InsecureSkipVerify bool
	URL                string
	Auth               string
	Org                string
}

// Parse loads env vars into e and validates them, returning an error if validation fails.
func (e *Env) Parse() error {
	e.Namespace = strings.TrimSpace(os.Getenv(WatchNamespaceEnvVar))
	e.Mode = strings.TrimSpace(os.Getenv("TYK_MODE"))
	e.URL = strings.TrimSpace(os.Getenv("TYK_URL"))
	e.Auth = strings.TrimSpace(os.Getenv("TYK_AUTH"))
	e.Org = strings.TrimSpace(os.Getenv("TYK_ORG"))
	e.InsecureSkipVerify, _ = strconv.ParseBool(os.Getenv("TYK_TLS_INSECURE_SKIP_VERIFY"))

	// verify
	sample := []struct {
		env, value string
	}{
		{"TYK_MODE", e.Mode},
		{"TYK_URL", e.URL},
		{"TYK_AUTH", e.Auth},
		{"TYK_ORG", e.Org},
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
	case "oss", "pro":
		return nil
	default:
		return fmt.Errorf("unknown TYK_MODE value %q", e.Mode)
	}
}

func (e *Env) JoinURL(parts ...string) string {
	return joinUrl(append([]string{e.URL}, parts...)...)
}

func joinUrl(parts ...string) string {
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

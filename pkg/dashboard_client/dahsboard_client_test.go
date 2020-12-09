package dashboard_client

import (
	"net/http"
	"testing"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
)

type Kase struct {
	Path    string
	Headers map[string]string
}

func compare(t *testing.T, k Kase, r *http.Request) {
	t.Helper()
	if k.Path != r.URL.Path {
		t.Errorf("path: expected %v got %v", k.Path, r.URL.Path)
	}
	for k, v := range k.Headers {
		x := r.Header.Get(k)
		if v != v {
			t.Errorf("headers %q: expected %q got %q", k, v, x)
		}
	}
}

func runKase(t *testing.T, e environmet.Env, fn func(*Client) error, kase Kase) {
}

package universal_client

import (
	"errors"
	"net/http"
	"testing"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Kase a single test case for an API call.
type Kase struct {
	Name    string
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
		if x != v {
			t.Errorf("headers %q: expected %q got %q", k, v, x)
		}
	}
}

// RunRequestKase this helps check if we are sending a correct request. This
// assumes fn will only perform a single API call, this ignores the response it
// only validates we are sending correct path/method/headers
func RunRequestKase(t *testing.T, e environmet.Env, fn func(Client) error, kase Kase) {
	t.Helper()
	t.Run(kase.Name, func(t *testing.T) {
		var request *http.Request
		x := Client{
			Env: e,
			Log: log.NullLogger{},
			Do: func(h *http.Request) (*http.Response, error) {
				request = h
				return nil, errors.New("TESTING")
			},
		}
		err := fn(x)
		if err != nil {
			t.Error(err)
			return
		}
		if request == nil {
			t.Error("no api call was made")
			return
		}
		compare(t, kase, request)
	})
}

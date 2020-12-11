package universal_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Kase a single test case for an API call.
type Kase struct {
	Name     string
	Request  RequestKase
	Response *ResponseKase
}

type RequestKase struct {
	Path     string
	Method   string
	Headers  map[string]string
	Callback func(*testing.T, *http.Request)
}

func (r *RequestKase) verify(t *testing.T, req *http.Request) {
	t.Helper()
	comparesString(t, "path", r.Path, req.URL.Path)
	comparesString(t, "method", r.Method, req.Method)
	compareHeaders(t, r.Headers, req.Header)
	if r.Callback != nil {
		r.Callback(t, req)
	}
}

type ResponseKase struct {
	Headers  map[string]string
	Body     string
	Callback func(*testing.T, *http.Response)
}

func (r *ResponseKase) verify(t *testing.T, res *http.Response, body string) {
	compareHeaders(t, r.Headers, res.Header)
	comparesString(t,
		"body",
		normalizeBody(t, r.Body),
		normalizeBody(t, body),
	)
	if r.Callback != nil {
		r.Callback(t, res)
	}
}

func normalizeBody(t *testing.T, body string) string {
	if body == "" {
		return body
	}
	o := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &o)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func comparesString(t *testing.T, field, expect, got string) {
	t.Helper()
	if got != expect {
		t.Errorf("%s: expected %s got %s", field, expect, got)
	}
}

func compareHeaders(t *testing.T, expect map[string]string, r http.Header) {
	t.Helper()
	for k, v := range expect {
		x := r.Get(k)
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
		var response *http.Response
		var body string
		var doErr error
		x := Client{
			Env: e,
			Log: log.NullLogger{},
			Do: func(h *http.Request) (*http.Response, error) {
				request = h
				if kase.Response != nil {
					response, doErr = Do(h)
					if doErr != nil {
						return nil, doErr
					}
					b, _ := ioutil.ReadAll(response.Body)
					response.Body.Close()
					body = string(b)
					response.Body = ioutil.NopCloser(bytes.NewReader(b))
					return response, nil
				}
				return nil, errors.New("TESTING")
			},
		}
		err := fn(x)
		if err != nil {
			t.Error(err)
			return
		}
		if doErr != nil {
			// something went wrong making upstream call
			t.Error(doErr)
			return
		}
		if request == nil {
			t.Error("no api call was made")
			return
		}
		kase.Request.verify(t, request)
		if response != nil {
			kase.Response.verify(t, response, body)
		}
	})
}

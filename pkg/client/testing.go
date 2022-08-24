package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	t.Helper()
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
func RunRequestKase(t *testing.T, e environmet.Env, fn func(context.Context) error, kase ...Kase) {
	t.Helper()

	var request []*http.Request
	var response []*http.Response
	var body []string
	var doErr []error

	x := Context{
		Env: e,
		Log: log.NullLogger{},
		Do: func(h *http.Request) (*http.Response, error) {
			request = append(request, h)

			res, err := Do(h)
			if err != nil {
				doErr = append(doErr, err)
				response = append(response, nil)
				return nil, err
			}

			doErr = append(doErr, nil)
			response = append(response, res)

			b, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, err
			}

			res.Body.Close()
			body = append(body, string(b))
			res.Body = io.NopCloser(bytes.NewReader(b))
			return res, nil
		},
	}

	err := fn(SetContext(context.Background(), x))
	if err != nil {
		t.Error(err)
		return
	}

	if request == nil {
		t.Error("no api call was made")
		return
	}

	if len(request) != len(kase) {
		t.Errorf("Mismatch expectations want %d got %d", len(kase), len(request))
		return
	}

	for i := 0; i < len(kase); i++ {
		k := kase[i]
		t.Run(k.Name, func(t *testing.T) {
			if doErr[i] != nil {
				t.Error(doErr[i])
				return
			}

			k.Request.verify(t, request[i])

			if k.Response != nil {
				if response[i] == nil {
					t.Error(doErr[i])
					return
				}
				k.Response.verify(t, response[i], body[i])
			}
		})
	}
}

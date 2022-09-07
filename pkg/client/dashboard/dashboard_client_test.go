package dashboard

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/client"
	"github.com/TykTechnologies/tyk-operator/pkg/client/universal"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
)

type (
	Kase         = client.Kase
	RequestKase  = client.RequestKase
	ResponseKase = client.ResponseKase
)

func env() environmet.Env {
	return environmet.Env{
		Environment: v1alpha1.Environment{
			Mode: "pro",
			URL:  "http://localhost:3000",
			Auth: "secret",
			Org:  "testing",
		},
	}
}

func newKlient() universal.Client {
	return &Client{}
}

// Sample loads sample file
func Sample(t *testing.T, name string, out interface{}) {
	x := filepath.Join("./samples/", name+".json")

	f, err := os.Open(x)
	if err != nil {
		t.Fatalf("%s: failed to open sample file %v", x, err)
	}

	defer f.Close()

	err = json.NewDecoder(f).Decode(out)
	if err != nil {
		t.Fatalf("%v: Failed td decode object ", x)
	}
}

func LoadSampleFile(t *testing.T, name string) *os.File {
	x := filepath.Join("./samples/", name)

	f, err := os.Open(x)
	if err != nil {
		t.Fatalf("%s: failed to open sample file %v", x, err)
	}

	return f
}

func ReadSample(t *testing.T, name string) string {
	f := LoadSampleFile(t, name+".json")

	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func ReadSampleFile(t *testing.T, name string) string {
	f := LoadSampleFile(t, name)

	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

type route struct {
	path    string
	body    string
	method  string
	headers map[string]string
	code    int
}

func (r *route) Serve(t *testing.T, w http.ResponseWriter) {
	code := r.code
	if code == 0 {
		code = http.StatusOK
	}

	for k, v := range r.headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(code)

	f := LoadSampleFile(t, r.body+".json")

	defer f.Close()

	io.Copy(w, f)
}

func mockDash(t *testing.T, r ...*route) http.Handler {
	xm := make(map[string][]*route)

	for _, v := range r {
		xm[v.method] = append(xm[v.method], v)
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if a, ok := xm[r.Method]; ok {
			for _, i := range a {
				if i.path == r.URL.Path {
					i.Serve(t, rw)
					return
				}
			}
		}
		rw.WriteHeader(http.StatusNotFound)
		f := LoadSampleFile(t, "404.json")
		defer f.Close()
		io.Copy(rw, f)
	})
}

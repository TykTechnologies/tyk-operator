package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/logr"
)

// ErrTODO is returned when a feature is not yet implemented
var ErrTODO = errors.New("TODO: This feature is not implemented yet")

// ErrNotFound is returned when an api call returns 404
var ErrNotFound = errors.New("Resource not found")

func IsTODO(err error) bool {
	return errors.Is(err, ErrTODO)
}

// IsNotFound returns true if err is ErrNotFound
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IgnoreNotFound returns nil if err is ErrNotFound
func IgnoreNotFound(err error) error {
	if !IsNotFound(err) {
		return err
	}
	return nil
}

var client = &http.Client{}

var once sync.Once

func init() {
	if os.Getenv(environmet.SkipVerify) == "true" {
		client.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
}

func JSON(res *http.Response, o interface{}) error {
	return json.NewDecoder(res.Body).Decode(o)
}

func Do(r *http.Request) (*http.Response, error) {
	return client.Do(r)
}

type HTTP struct {
	Env           environmet.Env
	Log           logr.Logger
	BeforeRequest func(*http.Request)
	Do            func(*http.Request) (*http.Response, error)
}

func (c HTTP) Environment() environmet.Env {
	return c.Env
}

func (c HTTP) Request(method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if c.BeforeRequest != nil {
		c.BeforeRequest(r)
	}
	return r, nil
}

func (c HTTP) JSON(method, url string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	b, err := v1alpha1.Marshal(body)
	if err != nil {
		return nil, err
	}
	fn = append(fn, AddHeaders(map[string]string{
		"Content-Type": "application/json",
	}))
	return c.Call(method, url, bytes.NewReader(b), fn...)
}

func (c HTTP) Get(url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return c.Call(http.MethodGet, url, body, fn...)
}

func (c HTTP) Post(url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return c.Call(http.MethodPost, url, body, fn...)
}

func (c HTTP) PostJSON(url string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	return c.JSON(http.MethodPost, url, body, fn...)
}

func (c HTTP) PutJSON(url string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	return c.JSON(http.MethodPut, url, body, fn...)
}

func (c HTTP) Delete(url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return c.Call(http.MethodDelete, url, body, fn...)
}

func (c HTTP) Call(method, url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	r, err := c.Request(method, url, body)
	if err != nil {
		return nil, err
	}
	for _, f := range fn {
		f(r)
	}
	var res *http.Response
	if c.Do != nil {
		res, err = c.Do(r)
	} else {
		res, err = client.Do(r)
	}
	values := []interface{}{
		"Method", method, "URL", url,
	}
	if res != nil {
		values = append(values, "Status", res.StatusCode)
	} else {
		values = append(values, "Status", err.Error())
	}
	values = append(values)
	c.Log.Info("Call", values...)
	if err == nil && res.StatusCode == http.StatusNotFound {
		res.Body.Close()
		return nil, ErrNotFound
	}
	return res, err
}

// AddQuery call back for adding url queries
func AddQuery(q map[string]string) func(*http.Request) {
	return func(h *http.Request) {
		query := h.URL.Query()
		for k, v := range q {
			query.Set(k, v)
		}
		h.URL.RawQuery = query.Encode()
	}
}

func AddHeaders(q map[string]string) func(*http.Request) {
	return func(h *http.Request) {
		for k, v := range q {
			h.Header.Add(k, v)
		}
	}
}

func SetHeaders(q map[string]string) func(*http.Request) {
	return func(h *http.Request) {
		for k, v := range q {
			h.Header.Set(k, v)
		}
	}
}

// Error dumps whole response plus body and return it as an error
func Error(res *http.Response) error {
	b, _ := ioutil.ReadAll(res.Body)
	return fmt.Errorf("%d API call failed with %v", res.StatusCode, string(b))
}

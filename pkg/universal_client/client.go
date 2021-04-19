package universal_client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
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
var clientInsecure = &http.Client{
	Transport: &http.Transport{
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
	},
}

func JSON(res *http.Response, o interface{}) error {
	return json.NewDecoder(res.Body).Decode(o)
}

func Do(r *http.Request) (*http.Response, error) {
	return client.Do(r)
}

type Client struct {
	Env environmet.Env
	Log logr.Logger
	Do  func(*http.Request) (*http.Response, error)
}

func BeforeRequest(r *http.Request, e environmet.Env) {
	r.Header.Set("content-type", "application/json")
	if e.Mode == "pro" {
		r.Header.Set("authorization", e.Auth)
	} else {
		r.Header.Set("x-tyk-authorization", e.Auth)
	}
}
func (c Client) Environment() environmet.Env {
	return c.Env
}

func (c Client) get(ctx context.Context) environmet.Env {
	if e := environmet.Get(ctx); !e.IsZero() {
		return e
	}
	return c.Env
}

func (c Client) Request(method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c Client) JSON(ctx context.Context, method string, url []string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	b, err := v1alpha1.Marshal(body)
	if err != nil {
		return nil, err
	}
	fn = append(fn, AddHeaders(map[string]string{
		"Content-Type": "application/json",
	}))
	return c.Call(ctx, method, url, bytes.NewReader(b), fn...)
}

func (c Client) Get(ctx context.Context, url []string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return c.Call(ctx, http.MethodGet, url, body, fn...)
}

func (c Client) Post(ctx context.Context, url []string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return c.Call(ctx, http.MethodPost, url, body, fn...)
}

func (c Client) PostJSON(ctx context.Context, url []string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	return c.JSON(ctx, http.MethodPost, url, body, fn...)
}

func (c Client) PutJSON(ctx context.Context, url []string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	return c.JSON(ctx, http.MethodPut, url, body, fn...)
}

func (c Client) Delete(ctx context.Context, url []string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return c.Call(ctx, http.MethodDelete, url, body, fn...)
}

func (c Client) Call(ctx context.Context, method string, url []string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	e := c.get(ctx)
	r, err := c.Request(method, e.JoinURL(url...), body)
	if err != nil {
		return nil, err
	}
	BeforeRequest(r, e)
	for _, f := range fn {
		f(r)
	}
	var res *http.Response
	if c.Do != nil {
		res, err = c.Do(r)
	} else {
		if e.InsecureSkipVerify {
			res, err = clientInsecure.Do(r)
		} else {
			res, err = client.Do(r)
		}
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

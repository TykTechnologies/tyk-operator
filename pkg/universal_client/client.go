package universal_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/logr"
)

var client = &http.Client{}

func JSON(res *http.Response, o interface{}) error {
	return json.NewDecoder(res.Body).Decode(o)
}

type Client struct {
	Env           environmet.Env
	Log           logr.Logger
	BeforeRequest func(*http.Request)
	Do            func(*http.Request) (*http.Response, error)
}

func (c Client) Request(method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if c.BeforeRequest != nil {
		c.BeforeRequest(r)
	}
	return r, nil
}

func (c Client) JSON(method, url string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	fn = append(fn, AddHeaders(map[string]string{
		"Content-Type": "application/json",
	}))
	return c.Call(method, url, bytes.NewReader(b), fn...)
}

func (c Client) Get(url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return c.Call(http.MethodGet, url, body, fn...)
}

func (c Client) Post(url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return c.Call(http.MethodPost, url, body, fn...)
}

func (c Client) PostJSON(url string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	return c.JSON(http.MethodPost, url, body, fn...)
}

func (c Client) PutJSON(url string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	return c.JSON(http.MethodPut, url, body, fn...)
}

func (c Client) Delete(url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return c.Call(http.MethodDelete, url, body, fn...)
}

func (c Client) Call(method, url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	r, err := c.Request(method, url, body)
	if err != nil {
		return nil, err
	}
	for _, f := range fn {
		f(r)
	}
	if c.Do != nil {
		return c.Do(r)
	}
	return client.Do(r)
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

// Error dumps whole response plus body and return it as an error
func Error(res *http.Response) error {
	b, _ := httputil.DumpResponse(res, true)
	return fmt.Errorf("API call failed with %v", string(b))
}

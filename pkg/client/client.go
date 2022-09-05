package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/logr"
)

// ErrTODO is returned when a feature is not yet implemented
var ErrTODO = errors.New("TODO: This feature is not implemented yet")

// ErrNotFound is returned when an api call returns 404
var (
	ErrNotFound = errors.New("Resource not found")
	ErrFailed   = errors.New("Failed api call")
)

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

func init() {
	if os.Getenv(v1alpha1.SkipVerify) == "true" {
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

// Context inforation needed to make a successful http api call
type Context struct {
	Env           environmet.Env
	Log           logr.Logger
	BeforeRequest func(*http.Request)
	Do            func(*http.Request) (*http.Response, error)
}

type contextKey struct{}

func SetContext(ctx context.Context, rctx Context) context.Context {
	return context.WithValue(ctx, contextKey{}, rctx)
}

func GetContext(ctx context.Context) Context {
	if c := ctx.Value(contextKey{}); c != nil {
		return c.(Context)
	}

	return Context{
		Log: logr.Discard(),
	}
}

func Request(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	BeforeRequest(r)

	return r, nil
}

func CallJSON(ctx context.Context, method, url string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	b, err := model.Marshal(body)
	if err != nil {
		return nil, err
	}

	fn = append(fn, AddHeaders(map[string]string{
		"Content-Type": "application/json",
	}))

	return Call(ctx, method, url, bytes.NewReader(b), fn...)
}

func Get(ctx context.Context, url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return Call(ctx, http.MethodGet, url, body, fn...)
}

func Post(ctx context.Context, url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return Call(ctx, http.MethodPost, url, body, fn...)
}

func PostJSON(ctx context.Context, url string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	return CallJSON(ctx, http.MethodPost, url, body, fn...)
}

func PutJSON(ctx context.Context, url string, body interface{}, fn ...func(*http.Request)) (*http.Response, error) {
	return CallJSON(ctx, http.MethodPut, url, body, fn...)
}

func Delete(ctx context.Context, url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	return Call(ctx, http.MethodDelete, url, body, fn...)
}

func Call(ctx context.Context, method, url string, body io.Reader, fn ...func(*http.Request)) (*http.Response, error) {
	rctx := GetContext(ctx)
	url = JoinURL(rctx.Env.URL, url)

	r, err := Request(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	for _, f := range fn {
		f(r)
	}

	var res *http.Response
	if rctx.Do != nil {
		res, err = rctx.Do(r)
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

	rctx.Log.Info("Call", values...)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()

		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		if len(b) > 0 {
			rctx.Log.Info(http.StatusText(res.StatusCode), "body", string(b))
		}

		switch res.StatusCode {
		case http.StatusNotFound:
			return nil, ErrNotFound
		default:
			return nil, ErrFailed
		}
	}

	return res, err
}

// AddQuery call back for adding url queries
func AddQuery(q url.Values) func(*http.Request) {
	return func(h *http.Request) {
		if len(q) == 0 {
			return
		}

		query := h.URL.Query()

		for k, v := range q {
			query[k] = append(query[k], v...)
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
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("%d API call failed with %v", res.StatusCode, string(b))
}

// JoinURL returns addition of  parts to the base e.URL
func JoinURL(base string, parts ...string) string {
	return Join(append([]string{base}, parts...)...)
}

func Join(parts ...string) string {
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

func BeforeRequest(r *http.Request) {
	ctx := GetContext(r.Context())
	r.Header.Set("content-type", "application/json")

	switch ctx.Env.Mode {
	case "pro":
		r.Header.Set("authorization", ctx.Env.Auth)
	case "ce":
		r.Header.Set("x-tyk-authorization", ctx.Env.Auth)
	}

	if ctx.BeforeRequest != nil {
		ctx.BeforeRequest(r)
	}
}

func LError(ctx context.Context, err error, msg string, kv ...interface{}) {
	GetContext(ctx).Log.Error(err, msg, kv...)
}

func LInfo(ctx context.Context, msg string, kv ...interface{}) {
	GetContext(ctx).Log.Info(msg, kv...)
}

func Result(res *http.Response, err error) (*model.Result, error) {
	var m model.Result

	if e := Data(&m)(res, err); e != nil {
		return nil, e
	}

	return &m, nil
}

func Data(o interface{}) func(*http.Response, error) error {
	return func(res *http.Response, err error) error {
		if err != nil {
			return err
		}

		defer res.Body.Close()

		return json.NewDecoder(res.Body).Decode(o)
	}
}

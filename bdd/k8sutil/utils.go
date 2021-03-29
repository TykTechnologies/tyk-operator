package k8sutil

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"moul.io/http2curl/v2"
)

// OpExpect expected outcome from k8s operation
type OpExpect uint

// ErrNotImplemented this is returned when a CB method is not supported yet
var ErrNotImplemented = errors.New("k8sutil: The method is not implemented")

// expected k8s op outcome for resources
const (
	Created OpExpect = iota + 1
	Configured
	Deleted
)

func (o OpExpect) String() string {
	switch o {
	case Created:
		return "created"
	case Configured:
		return "configured"
	case Deleted:
		return "deleted"
	default:
		return "Unknown"
	}
}

type CMD interface {
	Init(ns string) (cancel func() error, err error)
	CreateNS(ctx context.Context, ns string) error
	DeleteNS(ctx context.Context, ns string) error
	Create(ctx context.Context, file string, namespace string) error
	Configure(ctx context.Context, file string, namespace string) error
	Delete(ctx context.Context, file string, namespace string) error
}

var cmd CMD = CB{
	InitFN:      initFn,
	CreateFn:    create,
	CreateNSFn:  createNS,
	DeleteNSFn:  deleteNS,
	ConfigureFn: config,
	DeleteFn:    del,
}

func Init(ns, env string) (func() error, error) {
	switch env {
	case "ce", "pro":
	default:
		return nil, fmt.Errorf("Unknown TYK_ENV=%q", env)
	}
	return cmd.Init(ns)
}

// Create applies file to k8s cluster and ensure that thr resource is created
func Create(ctx context.Context, file string, namespace string) error {
	return cmd.Create(ctx, file, namespace)
}

func CreateNS(ctx context.Context, namespace string) error {
	return cmd.CreateNS(ctx, namespace)
}

func DeleteNS(ctx context.Context, namespace string) error {
	return cmd.DeleteNS(ctx, namespace)
}

// Delete applies file to k8s cluster and ensure that thr resource is deleted
func Delete(ctx context.Context, file string, namespace string) error {
	return cmd.Delete(ctx, file, namespace)
}

// Configure applies file to k8s cluster and ensure that thr resource is configured
func Configure(ctx context.Context, file string, namespace string) error {
	return cmd.Configure(ctx, file, namespace)
}

// CB is a helper struct satisfying CMD interface. Use the fields to provide
// callbacks for respective CMD method call.
type CB struct {
	InitFN      func(ns string) (func() error, error)
	CreateFn    func(ctx context.Context, file string, namespace string) error
	DeleteFn    func(ctx context.Context, file string, namespace string) error
	ConfigureFn func(ctx context.Context, file string, namespace string) error
	CreateNSFn  func(ctx context.Context, ns string) error
	DeleteNSFn  func(ctx context.Context, ns string) error
}

func (fn CB) Init(ns string) (func() error, error) {
	if fn.InitFN == nil {
		return nil, ErrNotImplemented
	}
	return fn.InitFN(ns)
}

func (fn CB) Create(ctx context.Context, file string, namespace string) error {
	if fn.CreateFn == nil {
		return ErrNotImplemented
	}
	return fn.CreateFn(ctx, file, namespace)
}

func (fn CB) CreateNS(ctx context.Context, namespace string) error {
	if fn.CreateNSFn == nil {
		return ErrNotImplemented
	}
	return fn.CreateNSFn(ctx, namespace)
}

func (fn CB) DeleteNS(ctx context.Context, namespace string) error {
	if fn.DeleteNSFn == nil {
		return ErrNotImplemented
	}
	return fn.DeleteNSFn(ctx, namespace)
}

func (fn CB) Delete(ctx context.Context, file string, namespace string) error {
	if fn.DeleteFn == nil {
		return ErrNotImplemented
	}
	return fn.DeleteFn(ctx, file, namespace)
}

func (fn CB) Configure(ctx context.Context, file string, namespace string) error {
	if fn.ConfigureFn == nil {
		return ErrNotImplemented
	}
	return fn.ConfigureFn(ctx, file, namespace)
}

func runCMD(cmd *exec.Cmd) (string, error) {
	a := fmt.Sprint(cmd.Args)
	fmt.Println("===> EXEC ", a)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed %s with %v : %s", a, err, string(output))
	}
	return string(output), nil
}

func create(ctx context.Context, file string, ns string) error {
	return expect(Created, Configured)(runFile(ctx, "apply", file, ns))
}

func createNS(ctx context.Context, ns string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "create", "ns", ns)
	o, err := runCMD(cmd)
	if err != nil {
		return err
	}
	if !strings.Contains(o, Created.String()) {
		return fmt.Errorf("k8sutil: expected ns to be %v got :%q", Created, o)
	}
	return nil
}

func deleteNS(ctx context.Context, ns string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "ns", ns)
	o, err := runCMD(cmd)
	if err != nil {
		return err
	}
	if !strings.Contains(o, Deleted.String()) {
		return fmt.Errorf("k8sutil: expected ns to be %v got :%q", Deleted, o)
	}
	return nil
}

func config(ctx context.Context, file string, ns string) error {
	return expect(Configured)(runFile(ctx, "apply", file, ns))
}

func del(ctx context.Context, file string, ns string) error {
	return expect(Deleted)(runFile(ctx, "delete", file, ns))
}

func runFile(ctx context.Context, op, file string, ns string) (OpExpect, error) {
	cmd := exec.CommandContext(ctx, "kubectl", op, "-f", file, "-n", ns)
	o, err := runCMD(cmd)
	if err != nil {
		return 0, err
	}
	for _, v := range []OpExpect{Created, Configured, Deleted} {
		if strings.Contains(o, v.String()) {
			return v, nil
		}
	}
	return 0, fmt.Errorf("k8sutil: unexpected output for cmd: %v output:%q", cmd.Args, o)
}

func expect(have ...OpExpect) func(OpExpect, error) error {
	return func(oe OpExpect, e error) error {
		if e != nil {
			return e
		}
		var h []string
		for _, v := range have {
			if v == oe {
				return nil
			}
			h = append(h, v.String())
		}
		return fmt.Errorf("expected %v got %v", h, oe)
	}
}

func initFn(ns string) (func() error, error) {
	return func() error { return nil }, setup(ns)
}

type writeFn func([]byte) (int, error)

func (fn writeFn) Write(b []byte) (int, error) {
	return fn(b)
}

var api TykAPI

func setup(ns string) error {
	label := "name=tyk"
	if os.Getenv("TYK_HELM_CHARTS") != "" {
		switch os.Getenv("TYK_MODE") {
		case "ce":
			label = "app=gateway-ce-tyk-headless"
			api.Container = "gateway-tyk-headless"
		case "pro":
			label = "app=gateway-pro-tyk-pro"
			api.Container = "gateway-tyk-pro"
		}
	} else {
		api.Container = "tyk"
	}
	e := exec.Command(
		"kubectl", "get", "pods", "-l", label, "-n", ns,
		"-o", "jsonpath={.items..metadata.name}",
	)
	o, err := e.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v:%v", err, string(o))
	}
	pod := string(bytes.TrimSpace(o))
	pod = strings.Split(pod, " ")[0]
	if pod == "" {
		return fmt.Errorf("failed to get tyk pod cmd=%v", e.Args)
	}
	api.Namespace = ns
	api.Pod = pod
	return nil
}

type TykAPI struct {
	Namespace, Pod, Container string
}

func Do(r *http.Request) (*http.Response, error) {
	return api.Do(r)
}

func unquote(s string) string {
	if s[0] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

const agent = "Go-http-client/1.1"

func (t TykAPI) Do(r *http.Request) (*http.Response, error) {
	c, err := http2curl.GetCurlCommand(r)
	if err != nil {
		return nil, err
	}
	cs := []string(*c)
	for i := 0; i < len(cs); i++ {
		cs[i] = unquote(cs[i])
	}
	e := exec.Command("kubectl", append(
		t.commands(), cs[1:]...,
	)...)
	if err != nil {
		return nil, err
	}
	fmt.Println(e.Args)
	var buf bytes.Buffer
	e.Stderr = os.Stderr
	e.Stdout = &buf
	if err := e.Run(); err != nil {
		return nil, err
	}
	res, err := http.ReadResponse(bufio.NewReader(&buf), r)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (t TykAPI) commands() []string {
	return []string{
		"exec", t.Pod, "-c", t.Container, "-n", t.Namespace,
		"--", "curl", "-s", "-i", "-A", agent, "-H", "Accept:",
	}
}

package k8sutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

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
	Got
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
	Init(ns, kubeconfig string) (cancel func() error, err error)
	CreateNS(ctx context.Context, ns string) error
	DeleteNS(ctx context.Context, ns string) error
	Create(ctx context.Context, file, namespace string) error
	Configure(ctx context.Context, file, namespace string) error
	Delete(ctx context.Context, file, namespace string) error
	Get(ctx context.Context, file, namespace string) error
}

var cmd CMD = CB{
	InitFN:      initFn,
	CreateFn:    create,
	CreateNSFn:  createNS,
	DeleteNSFn:  deleteNS,
	ConfigureFn: config,
	DeleteFn:    del,
	GetFn:       get,
}

func Init(ns, env, kubeconfig string) (func() error, error) {
	switch env {
	case "ce", "pro":
	default:
		return nil, fmt.Errorf("Unknown TYK_MODE=%q", env)
	}

	return cmd.Init(ns, kubeconfig)
}

// Create applies file to k8s cluster and ensure that thr resource is created
func Create(ctx context.Context, file, namespace string) error {
	return cmd.Create(ctx, file, namespace)
}

func CreateNS(ctx context.Context, namespace string) error {
	return cmd.CreateNS(ctx, namespace)
}

func DeleteNS(ctx context.Context, namespace string) error {
	return cmd.DeleteNS(ctx, namespace)
}

// Delete applies file to k8s cluster and ensure that thr resource is deleted
func Delete(ctx context.Context, file, namespace string) error {
	return cmd.Delete(ctx, file, namespace)
}

// Configure applies file to k8s cluster and ensure that thr resource is configured
func Configure(ctx context.Context, file, namespace string) error {
	return cmd.Configure(ctx, file, namespace)
}

func Get(ctx context.Context, file, namespace string) error {
	return cmd.Get(ctx, file, namespace)
}

// CB is a helper struct satisfying CMD interface. Use the fields to provide
// callbacks for respective CMD method call.
type CB struct {
	InitFN      func(ns, kubeconfig string) (func() error, error)
	CreateFn    func(ctx context.Context, file, namespace string) error
	DeleteFn    func(ctx context.Context, file, namespace string) error
	ConfigureFn func(ctx context.Context, file, namespace string) error
	GetFn       func(ctx context.Context, file, namespace string) error
	CreateNSFn  func(ctx context.Context, ns string) error
	DeleteNSFn  func(ctx context.Context, ns string) error
}

func (fn CB) Init(ns, kubeconfig string) (func() error, error) {
	if fn.InitFN == nil {
		return nil, ErrNotImplemented
	}

	return fn.InitFN(ns, kubeconfig)
}

func (fn CB) Create(ctx context.Context, file, namespace string) error {
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

func (fn CB) Delete(ctx context.Context, file, namespace string) error {
	if fn.DeleteFn == nil {
		return ErrNotImplemented
	}

	return fn.DeleteFn(ctx, file, namespace)
}

func (fn CB) Configure(ctx context.Context, file, namespace string) error {
	if fn.ConfigureFn == nil {
		return ErrNotImplemented
	}

	return fn.ConfigureFn(ctx, file, namespace)
}

func (fn CB) Get(ctx context.Context, file, namespace string) error {
	if fn.GetFn == nil {
		return ErrNotImplemented
	}

	return fn.GetFn(ctx, file, namespace)
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

func create(ctx context.Context, file, ns string) error {
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

func config(ctx context.Context, file, ns string) error {
	return expect(Configured)(runFile(ctx, "apply", file, ns))
}

func del(ctx context.Context, file, ns string) error {
	return expect(Deleted)(runFile(ctx, "delete", file, ns))
}

func get(ctx context.Context, file, ns string) error {
	return expect(Got)(runFile(ctx, "get", file, ns))
}

func runFile(ctx context.Context, op, file, ns string) (OpExpect, error) {
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

func initFn(ns, kubeconfig string) (func() error, error) {
	return func() error { return nil }, setup(ns, kubeconfig)
}

func initK8sPortForwardForPod(podName, podNamespace, kubeconfig string) (httpstream.Dialer, error) {
	c, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	roundTripper, upgrader, err := spdy.RoundTripperFor(c)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", podNamespace, podName)
	hostIP := strings.TrimLeft(c.Host, "https:/")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	return spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL), nil
}

var api TykAPI

func setup(ns, kubeconfig string) error {
	var label string

	mode := os.Getenv("TYK_MODE")

	switch mode {
	case "ce":
		label = "app=gateway-ce-tyk-headless"
		api.Container = "gateway-tyk-headless"
	case "pro":
		label = "app=gateway-pro-tyk-pro"
		api.Container = "gateway-tyk-pro"
	default:
		return fmt.Errorf("unknown mode %q", mode)
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
	api.Kubeconfig = kubeconfig

	dialer, err := initK8sPortForwardForPod(api.Pod, api.Namespace, api.Kubeconfig)
	if err != nil {
		return err
	}

	go func() {
		err = portForward(dialer)
		if err != nil {
			panic(err)
		}
	}()

	<-readyChan

	return nil
}

type TykAPI struct {
	Namespace  string
	Pod        string
	Container  string
	Kubeconfig string
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

var (
	stopChan, readyChan = make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut         = new(bytes.Buffer), new(bytes.Buffer)
)

func portForward(dialer httpstream.Dialer) error {
	fw, err := portforward.New(dialer, []string{"8080:8080"}, stopChan, readyChan, out, errOut)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}

func (t TykAPI) Do(r *http.Request) (*http.Response, error) {
	c, err := http2curl.GetCurlCommand(r)
	if err != nil {
		return nil, err
	}

	cs := []string(*c)

	for i := 0; i < len(cs); i++ {
		cs[i] = unquote(cs[i])
	}

	var hc http.Client
	resp, err := hc.Do(r)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()

			b, errResp := io.ReadAll(resp.Body)
			if errResp != nil {
				fmt.Println("failed to read the response body, err: ", errResp)
			} else {
				fmt.Println("failed to send the request, response: ", string(b))
			}
		}

		fmt.Println("failed to send the request, err: ", err)

		return nil, err
	}

	return resp, nil
}

func (t TykAPI) commands() []string {
	return []string{
		"exec", t.Pod, "-c", t.Container, "-n", t.Namespace,
		"--", "curl", "-s", "-i", "-A", agent, "-H", "Accept:",
	}
}

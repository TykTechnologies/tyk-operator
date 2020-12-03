package k8sutil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
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

func Init(ns string) (func() error, error) {
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
	InitFN      func(string) (func() error, error)
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
	return expect(Created)(runFile(ctx, "apply", file, ns))
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

func expect(have OpExpect) func(OpExpect, error) error {
	return func(oe OpExpect, e error) error {
		if e != nil {
			return e
		}
		if oe != have {
			return fmt.Errorf("expected %v got %v", have, oe)
		}
		return nil
	}
}

func initFn(ns string) (func() error, error) {
	comm := make(chan struct{})
	go set(comm, ns)
	select {
	case <-comm:
	case <-time.After(3 * time.Second):
		return nil, errors.New("Failed to setup port forwarding")
	}
	return func() error {
		comm <- struct{}{}
		select {
		case <-comm:
		case <-time.After(3 * time.Second):
			return errors.New("Failed to tear down port forwarding")
		}
		return nil
	}, nil
}

func set(comm chan struct{}, ns string) {
	for {
		kill, term, err := setup(ns)
		if err != nil {
			panic(err)
		}
		comm <- struct{}{}
		select {
		case <-term:
			kill()
			fmt.Println("===> reopening port forwarding")
			time.Sleep(time.Second)
		case <-comm:
			kill()
			comm <- struct{}{}
			return
		}
	}
}

type writeFn func([]byte) (int, error)

func (fn writeFn) Write(b []byte) (int, error) {
	return fn(b)
}

func setup(ns string) (func() error, chan struct{}, error) {
	// make sure we don't have the testing ns
	cmd := exec.Command("kubectl", "port-forward", "-n", ns, "svc/gw", "8000:8000")
	fmt.Println(cmd.Args)
	var once sync.Once
	firstLine := make(chan string, 1)
	fail := "failed to execute portforward in network namespace"
	term := make(chan struct{})
	cmd.Stderr = writeFn(func(b []byte) (int, error) {
		once.Do(func() { firstLine <- string(b) })
		if bytes.Contains(b, []byte(fail)) {
			term <- struct{}{}
		}
		return os.Stderr.Write(b)
	})
	cmd.Stdout = writeFn(func(b []byte) (int, error) {
		once.Do(func() { firstLine <- string(b) })
		return os.Stdout.Write(b)
	})
	err := cmd.Start()
	if err != nil {
		return nil, nil, err
	}
	ts := time.NewTimer(3 * time.Second)
	defer ts.Stop()
	select {
	case <-ts.C:
		return nil, nil, errors.New("timeout waiting for port forwarding")
	case b := <-firstLine:
		x := "Forwarding from 127.0.0.1:8000"
		if !strings.HasPrefix(b, x) {
			cmd.Process.Kill()
			return nil, nil, fmt.Errorf("expected %q got %q", x, b)
		}
	}
	return cmd.Process.Kill, term, nil
}

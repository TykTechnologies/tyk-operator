package k8sutil

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
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
	CreateNS(ctx context.Context, ns string) error
	DeleteNS(ctx context.Context, ns string) error
	Create(ctx context.Context, file string, namespace string) error
	Configure(ctx context.Context, file string, namespace string) error
	Delete(ctx context.Context, file string, namespace string) error
}

var cmd CMD = CB{
	CreateFn:    create,
	CreateNSFn:  createNS,
	DeleteNSFn:  deleteNS,
	ConfigureFn: config,
	DeleteFn:    del,
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
	CreateFn    func(ctx context.Context, file string, namespace string) error
	DeleteFn    func(ctx context.Context, file string, namespace string) error
	ConfigureFn func(ctx context.Context, file string, namespace string) error
	CreateNSFn  func(ctx context.Context, ns string) error
	DeleteNSFn  func(ctx context.Context, ns string) error
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
	if fn.DeleteFn == nil {
		return ErrNotImplemented
	}
	return fn.DeleteFn(ctx, file, namespace)
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

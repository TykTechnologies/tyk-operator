package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var (
	e       environmet.Env
	testenv env.Environment
)

const gatewaySVC = "tyk-operator-local-gateway-service"

type kubeConfigKey struct{}

func TestMain(t *testing.M) {
	e.Parse()
	testenv = env.New()
	testenv.Setup(func(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
		kubecfg, err := setupKind()
		if err != nil {
			return c1, err
		}
		client, err := klient.NewWithKubeConfigFile(kubecfg)
		if err != nil {
			return c1, err
		}
		c2.WithClient(client)
		return context.WithValue(c1, kubeConfigKey{}, kubecfg), nil
	}, func(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
		return c1, createGateway(c1, c2)
	},
	).Finish(func(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
		return c1, deleteGW(c1, c2)
	},
		func(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
			kubecfg := c1.Value(kubeConfigKey{}).(string)
			return c1, os.RemoveAll(kubecfg)
		})
	os.Exit(testenv.Run(t))
}

// createGateway creates a service that binds no nodeport
func createGateway(ctx context.Context, c2 *envconf.Config) error {
	var ls v1.ServiceList
	err := c2.Client().Resources(fmt.Sprintf("tyk%s-control-plane", string(e.Mode))).
		List(ctx, &ls)
	if err != nil {
		return err
	}
	for _, v := range ls.Items {
		if strings.HasPrefix(v.Name, "gateway") {
			o := v.Spec.Ports[0]
			s := v1.Service{}
			s.Name = gatewaySVC
			s.Namespace = v.Namespace
			s.Spec.Selector = v.Spec.Selector
			s.Spec.Type = v1.ServiceTypeNodePort
			s.Spec.Ports = []v1.ServicePort{
				{
					Port:       9000,
					TargetPort: o.TargetPort,
					NodePort:   31000,
				},
			}
			return c2.Client().Resources(v.Namespace).Create(ctx, &s)
		}
	}
	return nil
}

func deleteGW(ctx context.Context, c2 *envconf.Config) error {
	s := v1.Service{}
	s.Name = gatewaySVC
	s.Namespace = fmt.Sprintf("tyk%s-control-plane", string(e.Mode))
	return c2.Client().Resources().
		Delete(ctx, &s)
}

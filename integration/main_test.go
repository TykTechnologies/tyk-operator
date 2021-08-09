package integration

import (
	"context"
	"fmt"
	"log"
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

const (
	gatewaySVC       = "tyk-operator-local-gateway-service"
	adminSVC         = "tyk-operator-local-gateway-admin-service"
	gatewayLocalhost = "http://localhost:7000"
	adminLocalhost   = "http://localhost:7200"
)

type kubeConfigKey struct{}

func TestMain(t *testing.M) {
	e.Parse()
	if e.Mode == "" {
		log.Fatal("Missing TYK_MODE")
	}
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
		return c1, createLocalServices(c1, c2)
	},
	).Finish(func(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
		return c1, deleteLocalServices(c1, c2)
	},
		func(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
			kubecfg := c1.Value(kubeConfigKey{}).(string)
			return c1, os.RemoveAll(kubecfg)
		})
	os.Exit(testenv.Run(t))
}

// createLocalServices creates a service that binds no nodeport
func createLocalServices(ctx context.Context, c2 *envconf.Config) error {
	var ls v1.ServiceList
	err := c2.Client().Resources(envNS()).
		List(ctx, &ls)
	if err != nil {
		return err
	}
	var gwsvc, adminsvc *v1.Service
	for _, gw := range ls.Items {
		if strings.HasPrefix(gw.Name, "gateway") {
			gwsvc = &gw
		}
		if strings.HasPrefix(gw.Name, "dashboard") {
			adminsvc = &gw
		}
	}
	if isCE() {
		// for ce we use the same port for both admin and gateway
		adminsvc = gwsvc
	}
	return createServices(ctx, c2, gwsvc, adminsvc)
}

func deleteLocalServices(ctx context.Context, c2 *envconf.Config) error {
	for _, n := range []string{gatewaySVC, adminSVC} {
		s := v1.Service{}
		s.Name = n
		s.Namespace = envNS()
		if err := c2.Client().Resources().
			Delete(ctx, &s); err != nil {
			return err
		}
	}
	return nil
}

func envNS() string {
	return fmt.Sprintf("tyk%s-control-plane", e.Mode)
}

func createServices(ctx context.Context, c2 *envconf.Config, gw, admin *v1.Service) error {
	{
		// create gateway service
		o := gw.Spec.Ports[0]
		s := v1.Service{}
		s.Name = gatewaySVC
		s.Namespace = gw.Namespace
		s.Spec.Selector = gw.Spec.Selector
		s.Spec.Type = v1.ServiceTypeNodePort
		s.Spec.Ports = []v1.ServicePort{
			{
				Port:       9000,
				TargetPort: o.TargetPort,
				NodePort:   31000,
			},
		}
		err := c2.Client().Resources(gw.Namespace).Create(ctx, &s)
		if err != nil {
			return err
		}
	}
	// create admin service
	o := admin.Spec.Ports[0]
	s := v1.Service{}
	s.Name = adminSVC
	s.Namespace = gw.Namespace
	s.Spec.Selector = gw.Spec.Selector
	s.Spec.Type = v1.ServiceTypeNodePort
	s.Spec.Ports = []v1.ServicePort{
		{
			Port:       9900,
			TargetPort: o.TargetPort,
			NodePort:   31900,
		},
	}
	return c2.Client().Resources(gw.Namespace).Create(ctx, &s)
}

func isCE() bool {
	return e.Mode == "ce"
}

func isPro() bool {
	return e.Mode == "pro"
}

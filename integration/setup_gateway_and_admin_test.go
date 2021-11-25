package integration

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	gatewaySVC     = "tyk-operator-local-gateway-service"
	adminSVC       = "tyk-operator-local-gateway-admin-service"
	adminLocalhost = "http://localhost:7200"
)

func setupTyk(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
	return c1, createLocalServices(c1, c2)
}

func teardownTyk(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
	return c1, deleteLocalServices(c1, c2)
}

// createLocalServices creates a service that binds no nodeport
func createLocalServices(ctx context.Context, c2 *envconf.Config) error {
	var ls v1.ServiceList

	err := c2.Client().Resources(envNS()).
		List(ctx, &ls)
	if err != nil {
		return err
	}

	g := int(-1)
	a := int(-1)

	for k, gw := range ls.Items {
		if strings.HasPrefix(gw.Name, "gateway") {
			g = k
		}

		if strings.HasPrefix(gw.Name, "dashboard") {
			a = k
		}
	}

	if isCE() {
		a = g
	}

	if a == -1 || g == -1 {
		return errors.New("Failed to find tyk or dashboard service")
	}

	return createServices(ctx, c2, &ls.Items[g], &ls.Items[a])
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
	s.Namespace = admin.Namespace
	s.Spec.Selector = admin.Spec.Selector
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

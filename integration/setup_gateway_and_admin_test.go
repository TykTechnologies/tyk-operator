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

	operatorNamespace = "tyk-operator-system"
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

	list := []v1.Service{ls.Items[g], ls.Items[a]}

	return createServiceNode(ctx, c2, list)
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

func createServiceNode(ctx context.Context, c2 *envconf.Config, list []v1.Service) error {
	for index := range list {
		o := list[index].Spec.Ports[0]
		s := v1.Service{}

		s.Namespace = list[index].Namespace
		s.Spec.Selector = list[index].Spec.Selector
		s.Spec.Type = v1.ServiceTypeNodePort

		if strings.HasPrefix(list[index].Name, "gateway") {
			s.Name = gatewaySVC
			s.Spec.Ports = []v1.ServicePort{
				{
					Port:       9000,
					TargetPort: o.TargetPort,
					NodePort:   31000,
				},
			}
		} else {
			s.Name = adminSVC
			s.Spec.Ports = []v1.ServicePort{
				{
					Port:       9900,
					TargetPort: o.TargetPort,
					NodePort:   31900,
				},
			}
		}

		err := c2.Client().Resources(envNS()).Get(ctx, s.Name, list[index].Namespace, &s)
		// Create service if it doesn't exists
		if err != nil {
			err = c2.Client().Resources(list[index].Namespace).Create(ctx, &s)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
func isCE() bool {
	return e.Mode == "ce"
}

func isPro() bool {
	return e.Mode == "pro"
}

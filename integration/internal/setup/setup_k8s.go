package setup

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

type kubeConfigKey struct{}

func isKind() bool {
	kluster := os.Getenv("CLUSTER_NAME")
	if kluster == "" {
		kluster = "kind"
	}

	cmd := exec.Command("kind", "get", "clusters")

	var buf bytes.Buffer
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return false
	}

	return strings.Contains(buf.String(), kluster)
}

func kubeConf(o io.Writer) error {
	kluster := os.Getenv("CLUSTER_NAME")
	if kluster == "" {
		kluster = "kind"
	}

	cmd := exec.Command("kind", "get", "kubeconfig", "--name", kluster)
	cmd.Stdout = o
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func setupKind() (string, error) {
	if !isKind() {
		return "", errors.New("missing kind cluster")
	}

	f, err := os.CreateTemp("", "operator-kind-kubeconf")
	if err != nil {
		return "", err
	}

	if err := kubeConf(f); err != nil {
		f.Close()
		os.RemoveAll(f.Name())

		return "", err
	}

	defer f.Close()

	return f.Name(), nil
}

func Kubernetes(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
	kubecfg, err := setupKind()
	if err != nil {
		return c1, err
	}

	client, err := klient.NewWithKubeConfigFile(kubecfg)
	if err != nil {
		return c1, err
	}

	conf := client.RESTConfig()
	conf.ContentConfig.GroupVersion = &v1alpha1.GroupVersion
	conf.APIPath = "/apis"
	v1alpha1.AddToScheme(scheme.Scheme)

	conf.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	conf.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err = klient.New(conf)
	if err != nil {
		return c1, err
	}

	c2.WithClient(client)

	return context.WithValue(c1, kubeConfigKey{}, kubecfg), nil
}

func TeardownKubernetes(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
	kubecfg := c1.Value(kubeConfigKey{}).(string)
	return c1, os.RemoveAll(kubecfg)
}

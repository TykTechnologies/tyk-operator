package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/google/uuid"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	rand2 "k8s.io/apimachinery/pkg/util/rand"
	cr "sigs.k8s.io/controller-runtime/pkg/client"
	e2eKlient "sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	opNs                   = "tyk-operator-system"
	testSubGraphCRMetaName = "test-subgraph"
	testSubGraphSchema     = "test-schema"
	testSubGraphSDL        = "test-SDL"
	testApiDef             = "test-api-def"
	testOperatorCtx        = "mycontext"
	testSecurityPolicy     = "test-security-policy"
	gatewayLocalhost       = "http://localhost:7000"
	operatorSecret         = "tyk-operator-conf"
	tlsSecretCrtKey        = "tls.crt"
	tlsSecretKey           = "tls.key"
	mockTestMetaKey        = "mock_test"
)

// getUpdateCount returns update count stored in annotations of given object.
func getUpdateCount(object metav1.Object) (int, error) {
	annotations := object.GetAnnotations()
	if annotations == nil {
		return 0, fmt.Errorf(
			"failed to get annotations from %v/%v, nil annotations", object.GetName(), object.GetNamespace(),
		)
	}

	val, ok := annotations[mockTestMetaKey]
	if !ok {
		return 0, fmt.Errorf("key %v does not exist on annotations", mockTestMetaKey)
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("failed to convert annotation to integer, err: %v", err)
	}

	return valInt, nil
}

// mockVersion returns mock environment mode according to given environment mode. If the mode is "pro", it returns
// "mockdash"; otherwise, returns "mockgw".
func mockVersion(e *environmet.Env) v1alpha1.OperatorContextMode {
	if e.Mode == "pro" {
		return "mockdash"
	}

	return "mockgw"
}

// createTestClient creates controller-runtime client by wrapping given e2e test client. It can be used to create
// Reconciler for CRs such as ApiDefinitionReconciler.
func createTestClient(k e2eKlient.Client) (cr.Client, error) {
	scheme := runtime.NewScheme()

	cl, err := cr.New(k.RESTConfig(), cr.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return cl, v1alpha1.AddToScheme(scheme)
}

// generateApiDef generates a sample ApiDefinition CR. It won't create the newly created CR on your k8s.
func generateApiDef(ns string, mutateFn func(*v1alpha1.ApiDefinition)) *v1alpha1.ApiDefinition {
	var apiDef v1alpha1.ApiDefinition

	apiDef.Name = testApiDef
	apiDef.Namespace = ns
	apiDef.Spec.Name = testApiDef
	apiDef.Spec.Protocol = "http"
	apiDef.Spec.UseKeylessAccess = true
	apiDef.Spec.Active = true
	apiDef.Spec.VersionData = model.VersionData{
		DefaultVersion: "Default",
		NotVersioned:   true,
		Versions:       map[string]model.VersionInfo{"Default": {Name: "Default"}},
	}
	apiDef.Spec.Proxy = model.Proxy{
		ListenPath:      "/httpbin",
		TargetURL:       "http://httpbin.default.svc:8000",
		StripListenPath: true,
	}

	if mutateFn != nil {
		mutateFn(&apiDef)
	}

	return &apiDef
}

// generateSubGraphCR generates a sample SubGraph CR. It won't create the newly created CR on your k8s.
func generateSubGraphCR(namespace string, mutateFn func(graph *v1alpha1.SubGraph)) *v1alpha1.SubGraph {
	sg := &v1alpha1.SubGraph{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testSubGraphCRMetaName,
			Namespace: namespace,
		},
		Spec: v1alpha1.SubGraphSpec{
			SubGraphSpec: model.SubGraphSpec{
				Schema: testSubGraphSchema,
				SDL:    testSubGraphSDL,
			},
		},
	}

	if mutateFn != nil {
		mutateFn(sg)
	}

	return sg
}

// createTestAPIDef generates ApiDefinition CR and creates it on your cluster. You can modify new ApiDefinition
// CR via mutateFn.
func createTestAPIDef(ctx context.Context, c *envconf.Config, ns string, mutateFn func(*v1alpha1.ApiDefinition),
) (*v1alpha1.ApiDefinition, error) {
	apiDef := generateApiDef(ns, mutateFn)

	err := c.Client().Resources(ns).Create(ctx, apiDef)
	if err != nil {
		return nil, err
	}

	return apiDef, err
}

// waitForTykResourceCreation waits the creation of resource on Tyk GW or Dashboard. In order to understand whether
// the resource is created successfully or not, it checks status of CRs.
func waitForTykResourceCreation(envConf *envconf.Config, obj k8s.Object) error {
	err := wait.For(conditions.New(envConf.Client().Resources()).ResourceMatch(obj, func(obj k8s.Object) bool {
		switch val := obj.(type) {
		case *v1alpha1.ApiDefinition:
			if val.Status.ApiID != "" {
				return true
			}
		case *v1alpha1.SecurityPolicy:
			if val.Status.PolID != "" {
				return true
			}
		}

		return false
	}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

	return err
}

// createTestOperatorContext creates a sample OperatorContext resource on k8s.
func createTestOperatorContext(
	ctx context.Context,
	ns string,
	c *envconf.Config,
	fn func(operatorContext *v1alpha1.OperatorContext),
) (*v1alpha1.OperatorContext, error) {
	var operatorCtx v1alpha1.OperatorContext

	operatorCtx.Name = testOperatorCtx
	operatorCtx.Namespace = ns
	operatorCtx.Spec.FromSecret = &model.Target{
		Name:      operatorSecret,
		Namespace: opNs,
	}

	if fn != nil {
		fn(&operatorCtx)
	}

	return &operatorCtx, c.Client().Resources(ns).Create(ctx, &operatorCtx)
}

// createTestPolicy creates a test policy on k8s.
func createTestPolicy(ctx context.Context, c *envconf.Config, namespace string, mutateFn func(*v1alpha1.SecurityPolicy),
) (*v1alpha1.SecurityPolicy, error) {
	var policy v1alpha1.SecurityPolicy

	policy.Name = testSecurityPolicy + fmt.Sprintf("%d", rand2.Int())
	policy.Namespace = namespace
	policy.Spec = v1alpha1.SecurityPolicySpec{
		SecurityPolicySpec: model.SecurityPolicySpec{
			Name:   testSecurityPolicy + uuid.New().String(),
			Active: true,
			State:  "active",
		},
	}

	if mutateFn != nil {
		mutateFn(&policy)
	}

	err := c.Client().Resources(namespace).Create(ctx, &policy)
	if err != nil {
		return nil, err
	}

	return &policy, err
}

func genServerCertificate() ([]byte, []byte, error) {
	genCertificate := func(template *x509.Certificate) ([]byte, []byte, error) {
		priv, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			return nil, nil, err
		}

		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			return nil, nil, err
		}

		template.SerialNumber = serialNumber
		template.BasicConstraintsValid = true
		template.NotBefore = time.Now()
		template.NotAfter = template.NotBefore.Add(time.Hour)

		derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
		if err != nil {
			return nil, nil, err
		}

		var certPem, keyPem bytes.Buffer

		err = pem.Encode(&certPem, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		if err != nil {
			return nil, nil, err
		}

		err = pem.Encode(&keyPem, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
		if err != nil {
			return nil, nil, err
		}

		return certPem.Bytes(), keyPem.Bytes(), nil
	}

	certPem, privPem, err := genCertificate(&x509.Certificate{
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::")},
	})
	if err != nil {
		return nil, nil, err
	}

	return certPem, privPem, nil
}

// createTestTlsSecret creates a TLS type of Secret object on your k8s.
func createTestTlsSecret(ctx context.Context, ns string, c *envconf.Config, fn func(*v1.Secret)) (*v1.Secret, error) {
	var tlsSecret v1.Secret

	tlsSecret.Name = "test-tls-secret-name"
	tlsSecret.Namespace = ns
	tlsSecret.Data = make(map[string][]byte)

	certPem, privPem, err := genServerCertificate()
	if err != nil {
		return nil, err
	}

	tlsSecret.Type = v1.SecretTypeTLS
	tlsSecret.Data[tlsSecretKey] = privPem
	tlsSecret.Data[tlsSecretCrtKey] = certPem

	if fn != nil {
		fn(&tlsSecret)
	}

	return &tlsSecret, c.Client().Resources(ns).Create(ctx, &tlsSecret)
}

// generateEnvConfig creates a config structure to connect your Tyk installation. It parses k8s secret object
// and reads required connection credentials from there.
func generateEnvConfig(operatorConfSecret *v1.Secret) (environmet.Env, error) {
	data, ok := operatorConfSecret.Data["TYK_AUTH"]
	if !ok {
		return environmet.Env{}, errors.New("failed to parse TYK_AUTH from operator secret")
	}

	tykAuth := string(data)

	data, ok = operatorConfSecret.Data["TYK_ORG"]
	if !ok {
		return environmet.Env{}, errors.New("failed to parse TYK_ORG from operator secret")
	}

	tykOrg := string(data)
	tykVersion := "v4.2"

	data, ok = operatorConfSecret.Data["TYK_VERSION"]
	if ok && len(data) != 0 {
		tykVersion = string(data)
	}

	mode := os.Getenv("TYK_MODE")
	var tykConnectionURL string

	switch mode {
	case "pro":
		tykConnectionURL = adminLocalhost
	case "ce":
		tykConnectionURL = gatewayLocalhost
	}

	return environmet.Env{
		Environment: v1alpha1.Environment{
			Auth: tykAuth,
			Org:  tykOrg,
			Mode: v1alpha1.OperatorContextMode(mode),
			URL:  tykConnectionURL,
		},
		TykVersion: tykVersion,
	}, nil
}

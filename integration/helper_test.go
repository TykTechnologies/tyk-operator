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
	"time"

	"github.com/buger/jsonparser"

	"github.com/TykTechnologies/tyk-operator/api/model"
	"github.com/TykTechnologies/tyk-operator/api/v1alpha1"
	"github.com/TykTechnologies/tyk-operator/controllers"
	"github.com/TykTechnologies/tyk-operator/pkg/environment"
	"github.com/TykTechnologies/tyk-operator/pkg/keys"
	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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

	testOASCmName      = "test-oas-cm"
	testOASConfKeyName = "test-oas.json"
	testOASCrdName     = "test-oas"
	testOASDoc         = `{
	"info": {
	  "title": "Petstore",
	  "version": "1.0.0"
	},
	"openapi": "3.0.3",
	"components": {},
	"paths": {},
	"x-tyk-api-gateway": {
	  "info": {
		"name": "Petstore",
		"state": {
		  "active": true
		}
	  },
	  "upstream": {
		"url": "https://petstore.swagger.io/v2"
	  },
	  "server": {
		"authentication": {
		  "baseIdentityProvider": "auth_token",
          "enabled": true
		},
		"listenPath": {
		  "value": "/petstore/",
		  "strip": true
		}
	  }
	}
  }`
)

// createTestClient creates controller-runtime client by wrapping given e2e test client. It can be used to create
// Reconciler for CRs such as ApiDefinitionReconciler.
func createTestClient(k e2eKlient.Client) (cr.Client, error) {
	scheme := runtime.NewScheme()

	cl, err := cr.New(k.RESTConfig(), cr.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	if err := v1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := networkingv1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return cl, v1alpha1.AddToScheme(scheme)
}

// generateApiDef generates a sample ApiDefinition CR. It won't create the newly created CR on your k8s.
func generateApiDef(ns string, mutateFn func(*v1alpha1.ApiDefinition)) *v1alpha1.ApiDefinition {
	var apiDef v1alpha1.ApiDefinition

	useKeylessAccess := true
	active := true
	stripListenPath := true
	listenPath := "httpbin"

	apiDef.Name = testApiDef
	apiDef.Namespace = ns
	apiDef.Spec.Name = testApiDef
	apiDef.Spec.Protocol = "http"
	apiDef.Spec.UseKeylessAccess = &useKeylessAccess
	apiDef.Spec.Active = &active
	apiDef.Spec.VersionData = model.VersionData{
		DefaultVersion: "Default",
		NotVersioned:   true,
		Versions:       map[string]model.VersionInfo{"Default": {Name: "Default"}},
	}
	apiDef.Spec.Proxy = model.Proxy{
		ListenPath:      &listenPath,
		TargetURL:       "http://httpbin.default.svc:8000",
		StripListenPath: &stripListenPath,
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
		case *v1alpha1.TykOasApiDefinition:
			if val.Status.ID != "" {
				return true
			}
		}

		return false
	}), wait.WithTimeout(defaultWaitTimeout), wait.WithInterval(defaultWaitInterval))

	return err
}

// createTestOperatorContext creates a sample OperatorContext resource on k8s.
func createTestOperatorContext(ctx context.Context, ns string, c *envconf.Config) (*v1alpha1.OperatorContext, error) {
	var operatorCtx v1alpha1.OperatorContext

	secret_ns := opNs
	operatorCtx.Name = testOperatorCtx
	operatorCtx.Namespace = ns
	operatorCtx.Spec.FromSecret = &model.Target{
		Name:      operatorSecret,
		Namespace: &secret_ns,
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
func generateEnvConfig(ctx context.Context, envConf *envconf.Config) (environment.Env, error) {
	operatorConfSecret := v1.Secret{}

	err := envConf.Client().Resources(opNs).Get(ctx, operatorSecret, opNs, &operatorConfSecret)
	if err != nil {
		return environment.Env{}, err
	}

	data, ok := operatorConfSecret.Data["TYK_AUTH"]
	if !ok {
		return environment.Env{}, errors.New("failed to parse TYK_AUTH from operator secret")
	}

	tykAuth := string(data)

	data, ok = operatorConfSecret.Data["TYK_ORG"]
	if !ok {
		return environment.Env{}, errors.New("failed to parse TYK_ORG from operator secret")
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

	return environment.Env{
		Environment: v1alpha1.Environment{
			Auth: tykAuth,
			Org:  tykOrg,
			Mode: v1alpha1.OperatorContextMode(mode),
			URL:  tykConnectionURL,
		},
		TykVersion: tykVersion,
	}, nil
}

func createTestOASApi(ctx context.Context, ns string, c *envconf.Config, tykOASDoc string, oasLabels map[string]string,
) (*v1alpha1.TykOasApiDefinition, *v1.ConfigMap, error) {
	cm := &v1.ConfigMap{}
	cm.Name = testOASCmName
	cm.Namespace = ns

	if tykOASDoc == "" {
		cm.Data = map[string]string{testOASConfKeyName: testOASDoc}
	} else {
		cm.Data = map[string]string{testOASConfKeyName: tykOASDoc}
	}

	if err := c.Client().Resources(ns).Create(ctx, cm); err != nil {
		return nil, nil, err
	}

	if err := wait.For(conditions.New(c.Client().Resources(ns)).ResourceMatch(cm, func(object k8s.Object) bool {
		conf := object.(*v1.ConfigMap) //nolint:errcheck

		return conf.UID != ""
	})); err != nil {
		return nil, nil, err
	}

	tykOAS := &v1alpha1.TykOasApiDefinition{}

	tykOAS.Name = testOASCrdName
	tykOAS.Namespace = ns
	tykOAS.Spec.TykOAS.ConfigmapRef = v1alpha1.ConfigMapReference{
		Name:      cm.Name,
		Namespace: ns,
		KeyName:   testOASConfKeyName,
	}

	if oasLabels != nil {
		tykOAS.ObjectMeta.SetLabels(oasLabels)
	}

	err := c.Client().Resources(ns).Create(ctx, tykOAS)

	return tykOAS, cm, err
}

/*
Helpers for Ingress integration tests
*/

// expectedOasStatus represents a struct to describe expectations from TykOasApiDefinition CR status
type expectedOasStatus struct {
	upstreamURL string
	status      v1alpha1.TransactionStatus
	listenPath  string
}

// checkOasApiStatus checks if the OAS API status satisfies the expected structure
func checkOasApiStatus(oasApi *v1alpha1.TykOasApiDefinition, expectations expectedOasStatus) error {
	if len(oasApi.Status.ID) == 0 {
		return fmt.Errorf("invalid Status ID, status ID must be updated")
	}

	if oasApi.Status.TargetURL != expectations.upstreamURL {
		return fmt.Errorf("invalid TargetURL in TykOasApiDefinition status, TykOasApiDefinition Status: %+v", oasApi.Status)
	}

	if oasApi.Status.LatestTransaction.Status != expectations.status {
		return fmt.Errorf(
			"invalid LatestTransaction.Status in TykOasApiDefinition status, TykOasApiDefinition Status: %+v",
			oasApi.Status,
		)
	}

	if oasApi.Status.ListenPath != expectations.listenPath {
		return fmt.Errorf(
			"invalid ListenPath in TykOasApiDefinition status, TykOasApiDefinition Status: %+v", oasApi.Status,
		)
	}

	return nil
}

// checkOasApisReferExistingConfigMaps checks if each member of oasApis refers to the existing configMaps.
func checkOasApisReferExistingConfigMaps(oasApis []v1alpha1.TykOasApiDefinition, configMaps []v1.ConfigMap) bool {
	type mapKey string
	key := func(ns, name string) mapKey {
		return mapKey(ns + "/" + name)
	}

	configMapSet := make(map[mapKey]bool)
	for i := range configMaps {
		configMapSet[key(configMaps[i].Namespace, configMaps[i].Name)] = true
	}

	for i := range oasApis {
		k := key(oasApis[i].Spec.TykOAS.ConfigmapRef.Namespace, oasApis[i].Spec.TykOAS.ConfigmapRef.Name)
		if _, exists := configMapSet[k]; !exists {
			return false
		}
	}

	return true
}

// validateOasAuthentication checks for if authentication related configurations are inherited from templates.
// It takes two byte arrays; the first represents OAS API returned from Tyk Client and the second represents
// template ConfigMap's data.
func validateOasAuthentication(tykOasApi, tplCmData []byte) error {
	tykAuthType, err := jsonparser.GetString(tykOasApi, controllers.ServerAuthenticationBaseIdentityProviderKeys...)
	if err != nil {
		return fmt.Errorf("failed to parse 'server.authentication' from OasApiDefinition fetched from Tyk, err: %v", err)
	}

	expectedAuthType, err := jsonparser.GetString(tplCmData, controllers.ServerAuthenticationBaseIdentityProviderKeys...)
	if err != nil {
		return fmt.Errorf("failed to parse 'info.name' from  fetched from Tyk, err: %v", err)
	}

	if tykAuthType != expectedAuthType {
		return fmt.Errorf("unexpected OAS Auth Type, expected %v got %v", expectedAuthType, tykAuthType)
	}

	return nil
}

// validateStrField validates if the given byte array, which represents OAS API from Tyk in JSON format, has
// expected info field at JSON keys field.
func validateStrField(tykOasApi []byte, expectedVal string, keys []string) error {
	v, err := jsonparser.GetString(tykOasApi, keys...)
	if err != nil {
		return fmt.Errorf("failed to parse %+v from OasApiDefinition fetched from Tyk, err: %v", keys, err)
	}

	if v != expectedVal {
		return fmt.Errorf("unexpected OAS API spec at field: %v, expected %v got %v", keys, expectedVal, v)
	}

	return nil
}

// validateDomainName checks if the given domain name is created accordingly in the given tykOasApi string which
// represents Tyk OAS API definition fetched from Tyk in JSON format.
func validateDomainName(tykOasApi []byte, statusDomain string) error {
	if statusDomain == "" {
		return nil
	}

	customDomainName, err := jsonparser.GetString(tykOasApi, controllers.ServerCustomDomainNameKeys...)
	if err != nil {
		return fmt.Errorf(
			"failed to parse 'customDomain.name' from OasApiDefinition fetched from Tyk, err: %v",
			err,
		)
	}

	if statusDomain != customDomainName {
		return fmt.Errorf("unexpected OAS API CustomDomain.name, expected %v got %v",
			statusDomain, customDomainName,
		)
	}

	customDomainEnabled, err := jsonparser.GetBoolean(tykOasApi, controllers.ServerCustomDomainEnabledKeys...)
	if err != nil {
		return fmt.Errorf(
			"failed to parse 'customDomain.enabled' from OasApiDefinition fetched from Tyk, err: %v", err,
		)
	}

	if !customDomainEnabled {
		return fmt.Errorf("unexpected OAS API CustomDomain.enabled, expected %v got %v",
			true, customDomainEnabled)
	}

	return nil
}

// ingTplMeta represents a struct holding metadata information of the template given to Ingress.
// The name field represents 'tyk.io/template' label,
// while kind field represents 'tyk.io/template-kind' label.
type ingTplMeta struct {
	name string
	kind string
}

// newIngress creates a new Ingress object with the specified configurations.
// Based on the configurations, newIngress creates two rules; one with the host and another one is without host.
func newIngress(tpl ingTplMeta, ingName, ingNs, host, path, svcName string, svcPort int32) *networkingv1.Ingress {
	annotations := map[string]string{
		"kubernetes.io/ingress.class":         "tyk",
		keys.TykOasApiDefinitionTemplateLabel: tpl.name,
	}
	if tpl.kind != "" {
		annotations[keys.IngressTemplateKindAnnotation] = tpl.kind
	}

	pathType := networkingv1.PathTypePrefix

	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingName,
			Namespace:   ingNs,
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     path,
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: svcName,
											Port: networkingv1.ServiceBackendPort{
												Number: svcPort,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     path,
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: svcName,
											Port: networkingv1.ServiceBackendPort{
												Number: svcPort,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

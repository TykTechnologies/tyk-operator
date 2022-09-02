package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt"
)

var preloadImagesList = []struct{ name, image, version string }{
	{"mongo", "mongo", "latest"},
	{"rbac", "gcr.io/kubebuilder/kube-rbac-proxy", "v0.8.0"},
	{"redis", "k8s.gcr.io/redis", "e2e"},
	{"httpbin", "docker.io/kennethreitz/httpbin", ""},
	{"gateway", "tykio/tyk-gateway", "v3.1.2"},
	{"dash", "tykio/tyk-dashboard", "v3.2.1"},
	{"bash", "bash", "5.1"},
	{"busybox", "busybox", "1.32"},
	{"grpc", "mangomm/tyk-grpc-plugin", ""},
	{"cert-manager-cainjector", "quay.io/jetstack/cert-manager-cainjector", "v1.3.1"},
	{"cert-manager-controller", "quay.io/jetstack/cert-manager-controller", "v1.3.1"},
	{"cert-manager-webhook", "quay.io/jetstack/cert-manager-webhook", "v1.3.1"},
}

// Config configuration for booting operator environment
type Config struct {
	WorkDir       string
	PreloadImages bool
	Tyk           Tyk
	Operator      Operator
}

// Chart returns path to the helm chart to install
func (c Config) Chart() string {
	return filepath.Join(c.Tyk.Charts, chartDir())
}

// Values returns path to values.yaml used to install the chart
func (c Config) Values() string {
	return filepath.Join(c.WorkDir, "helm", chartDir(), "values.yaml")
}

func (c *Config) bind(mode string) {
	// set working directory
	wd, err := os.Getwd()
	if err != nil {
		exit(err)
	}

	c.WorkDir = filepath.Join(wd, "ci")
	env(&c.WorkDir, "TYK_OPERATOR_WORK_DIR")

	var strPreloadImages string
	env(&strPreloadImages, "TYK_OPERATOR_PRELOAD_IMAGES")

	c.PreloadImages, err = strconv.ParseBool(strPreloadImages)
	if err != nil {
		// set default value to false
		c.PreloadImages = false
	}

	c.Tyk.bind(c.WorkDir, mode)
	c.Operator.bind(mode)
}

type Tyk struct {
	License     string
	Auth        string
	Org         string
	Mode        string
	URL         string
	AdminSecret string
	Namespace   string
	Charts      string
}

func (t *Tyk) ce() {
	t.Mode = "ce"
	t.URL = "http://tyk.tykce-control-plane.svc.cluster.local:8001"
	t.Auth = "foo"
	t.Org = ""
	t.Namespace = "tykce-control-plane"
}

func (t *Tyk) helm(workdir string) {
	t.Charts = filepath.Join(workdir, "charts")

	if t.Mode == "pro" {
		t.URL = "http://dashboard-svc-ci-tyk-pro.tykpro-control-plane.svc.cluster.local:3000"
	} else {
		t.URL = "http://gateway-svc-ce-tyk-headless.tykce-control-plane.svc.cluster.local:8080"
	}
}

func (t *Tyk) pro() {
	t.Mode = "pro"
	t.URL = "http://dashboard.tykpro-control-plane.svc.cluster.local:3000"
	t.Namespace = "tykpro-control-plane"
	t.AdminSecret = "54321"
}

func (t *Tyk) bind(workdir, mode string) {
	t.Mode = mode

	if mode == "pro" {
		t.pro()
	} else {
		t.ce()
	}

	t.helm(workdir)
	env(&t.URL, "TYK_URL")
	env(&t.Org, "TYK_ORG")
	env(&t.Auth, "TYK_AUTH")
	env(&t.Namespace, "TYK_NAMESPACE")
	env(&t.AdminSecret, "TYK_ADMIN_SECRET")
	env(&t.Charts, "TYK_HELM_CHARTS")
	env(&t.License, "TYK_DB_LICENSEKEY")
}

func env(dest *string, name string) {
	if a := os.Getenv(name); a != "" {
		*dest = a
	}
}

type Operator struct {
	Namespace            string
	SecretName           string
	CertManager          string
	CertManagerNamespace string
}

func (o *Operator) defaults() {
	o.Namespace = "tyk-operator-system"
	o.SecretName = "tyk-operator-conf"
	o.CertManager = "https://github.com/cert-manager/cert-manager/releases/download/v1.8.0/cert-manager.yaml"
	o.CertManagerNamespace = "cert-manager"
}

func (o *Operator) bind(mode string) {
	o.defaults()
	env(&o.Namespace, "TYK_OPERATOR_NAMESPACE")
	env(&o.SecretName, "TYK_OPERATOR_SECRET_NAME")
	env(&o.CertManager, "TYK_OPERATOR_CERT_MANAGER")
	env(&o.CertManagerNamespace, "TYK_OPERATOR_CERT_MANAGER_NAMESPACE")
}

var config Config

var (
	mode       = flag.String("mode", os.Getenv("TYK_MODE"), "ce for community and pro for pro")
	debug      = flag.Bool("debug", false, "prints lots of details")
	cluster    = flag.String("cluster", "", "cluster name") //nolint
	tykVersion = flag.String("tyk_version", "v4.0", "tyk version to test against")
)

func chartDir() string {
	switch config.Tyk.Mode {
	case "ce":
		return "tyk-headless"
	case "pro":
		return "tyk-pro"
	default:
		return ""
	}
}

func deployDir() string {
	switch config.Tyk.Mode {
	case "ce":
		return "tyk-ce"
	case "pro":
		return "tyk-pro"
	default:
		return ""
	}
}

func main() {
	flag.Parse()
	config.bind(*mode)

	if config.PreloadImages {
		preloadImages()
	}

	submodule()
	createNamespaces()
	common()
	installTykStack()
	operator()
}

func submodule() {
	say("Setup helm charts submodule ...")

	cmd := exec.Command("git", "submodule", "init")
	cmd.Stderr = os.Stderr

	if *debug {
		cmd.Stdout = os.Stdout
	}

	exit(cmd.Run())
	cmd = exec.Command("git", "submodule", "update")
	cmd.Stderr = os.Stderr

	if *debug {
		cmd.Stdout = os.Stdout
	}

	exit(cmd.Run())
	ok()
}

func pro(fn func()) {
	if config.Tyk.Mode == "pro" {
		fn()
	}
}

func k(args ...string) error {
	return kf(func(c *exec.Cmd) {
		if *debug {
			c.Stderr = os.Stderr
			c.Stdout = os.Stdout
		}
	}, args...)
}

func kl(args ...string) error {
	return kf(func(c *exec.Cmd) {
		c.Stderr = os.Stderr
	}, args...)
}

func kf(fn func(*exec.Cmd), args ...string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Dir = config.WorkDir
	fn(cmd)

	if *debug {
		fmt.Println("==>", cmd.Args)
	}

	return cmd.Run()
}

// Create all namespaces needed to run ci for the operator. Only two namespaces
// are needed one is where tyk control plane will be deployed and the second one
// is where the operator will be deployed.
//
// They will only be created if they don't exist yet so it is safe to run this
// multiple times
func createNamespaces() {
	say("Creating Namespaces ...")

	if !hasNS(config.Tyk.Namespace) {
		exit(kl("create", "namespace", config.Tyk.Namespace))
	}

	if !hasNS(config.Operator.Namespace) {
		exit(kl("create", "namespace", config.Operator.Namespace))
	}

	ok()
}

func createSecret() {
	say("Creating Secret ... ")

	if !hasOperatorSecret() {
		pro(func() {
			var buf bytes.Buffer
			exit(kf(func(c *exec.Cmd) {
				c.Stdout = &buf
			}, "get", "secret",
				config.Operator.SecretName, "-n", config.Tyk.Namespace, "-o", "json"))
			o := struct {
				Data map[string]string `json:"data"`
			}{}
			exit(json.Unmarshal(buf.Bytes(), &o))
			for k, v := range o.Data {
				x, _ := base64.StdEncoding.DecodeString(v)
				o.Data[k] = string(x)
			}
			config.Tyk.Auth = o.Data["TYK_AUTH"]
			config.Tyk.Org = o.Data["TYK_ORG"]
			config.Tyk.Mode = o.Data["TYK_MODE"]
			config.Tyk.URL = o.Data["TYK_URL"]
		})

		exit(kl("create", "secret",
			"-n", config.Operator.Namespace,
			"generic", config.Operator.SecretName,
			"--from-literal", fmt.Sprintf("TYK_AUTH=%s", config.Tyk.Auth),
			"--from-literal", fmt.Sprintf("TYK_ORG=%s", config.Tyk.Org),
			"--from-literal", fmt.Sprintf("TYK_MODE=%s", config.Tyk.Mode),
			"--from-literal", fmt.Sprintf("TYK_URL=%s", config.Tyk.URL),
		))
	}

	ok()
}

func ok() {
	sayn("ok")
}

func say(a ...interface{}) {
	fmt.Print(a...)
}

func sayn(a ...interface{}) {
	fmt.Println(a...)
}

func exit(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func createRedis() {
	say("Creating Redis ....")

	if !hasDeployment("deployment/redis", config.Tyk.Namespace) {
		f := filepath.Join(config.WorkDir, deployDir(), "redis")
		exit(kl(
			"apply", "-f", f,
			"-n", config.Tyk.Namespace,
		))
		ok()
		say("Waiting for redis to be ready ...")
		exit(kl(
			"rollout", "status", "deployment/redis", "-n", config.Tyk.Namespace,
		))
	}

	ok()
}

func createMongo() {
	say("Creating Mongo ....")

	if !hasDeployment("deployment/mongo", config.Tyk.Namespace) {
		f := filepath.Join(config.WorkDir, deployDir(), "mongo")
		exit(kl(
			"apply", "-f", f,
			"-n", config.Tyk.Namespace,
		))
		ok()
		say("Waiting for mongo to be ready ...")
		exit(kl(
			"rollout", "status", "deployment/mongo", "-n", config.Tyk.Namespace,
		))
	}

	ok()
}

func installTykStack() {
	say("Installing helm chart ...")

	if !hasTykChart() {
		if config.Tyk.Mode == "pro" {
			if config.Tyk.License == "" {
				exit(errors.New("Dashboard license is empty"))
			}

			token, _, err := new(jwt.Parser).ParseUnverified(config.Tyk.License, jwt.MapClaims{})
			if err != nil {
				exit(err)
			}

			c := token.Claims.(jwt.MapClaims)
			if err := c.Valid(); err != nil {
				exit(err)
			}
		}

		cmd := exec.Command("helm", "install", config.Tyk.Mode,
			"-f", config.Values(),
			config.Chart(),
			"--set", fmt.Sprintf("dash.license=%s", config.Tyk.License),
			"--set", fmt.Sprintf("dash.image.tag=%s", *tykVersion),
			"--set", fmt.Sprintf("gateway.image.tag=%s", *tykVersion),
			"-n", config.Tyk.Namespace,
			"--wait",
		)
		cmd.Stderr = os.Stderr

		if *debug {
			cmd.Stdout = os.Stdout
			fmt.Println(cmd.Args)
		}

		exit(cmd.Run())
	}

	ok()
}

func hasTykChart() bool {
	var buf bytes.Buffer

	cmd := exec.Command("helm", "list",
		"-n", config.Tyk.Namespace,
	)
	cmd.Dir = config.WorkDir
	cmd.Stdout = &buf
	exit(cmd.Run())

	return strings.Contains(buf.String(), config.Tyk.Mode)
}

func hasNS(name string) bool {
	return k("get", "ns", name) == nil
}

func hasOperatorSecret() bool {
	return k("get", "secret", "-n", config.Operator.Namespace, config.Operator.SecretName) == nil
}

func hasDeployment(depName, namespace string) bool {
	return k("get", depName, "-n", namespace) == nil
}

func createCertManager() {
	say("Installing cert-manager ...")

	if !hasDeployment("deployment/cert-manager", config.Operator.CertManagerNamespace) {
		exit(kl(
			"apply",
			"--validate=false",
			"-f", config.Operator.CertManager,
		))
		ok()
		say("Waiting for cert-manager to be ready ...")
		exit(kl(
			"rollout", "status", "deployment/cert-manager", "-n", config.Operator.CertManagerNamespace,
		))
		exit(kl(
			"rollout", "status", "deployment/cert-manager-cainjector", "-n", config.Operator.CertManagerNamespace,
		))
		exit(kl(
			"rollout", "status", "deployment/cert-manager-webhook", "-n", config.Operator.CertManagerNamespace,
		))
		ok()
	}

	ok()
}

func common() {
	deployHTTPBIN()
	deployGRPCPlugin()
	createRedis()
	pro(createMongo)
}

func operator() {
	createCertManager()
	createSecret()
}

func deployHTTPBIN() {
	say("Deploying httpbin ...")

	if !hasDeployment("deployment/httpbin", "default") {
		exit(k("apply", "-f", filepath.Join(config.WorkDir, "upstreams")))
		ok()
		say("Waiting for httpbin to be ready ...")
		exit(kl(
			"rollout", "status", "--timeout", "2m", "deployment/httpbin",
		))
	}

	ok()
}

func deployGRPCPlugin() {
	say("Deploying grpc-plugin ...")

	if !hasDeployment("deployment/grpc-plugin", config.Tyk.Namespace) {
		exit(k("apply", "-f", filepath.Join(config.WorkDir, "grpc-plugin"), "-n", config.Tyk.Namespace))
		ok()
		say("Waiting for grpc-plugin to be ready ...")
		exit(kl(
			"rollout", "status", "deployment/grpc-plugin", "-n", config.Tyk.Namespace,
		))
	}

	ok()
}

func isKind() bool {
	var buf bytes.Buffer

	cmd := exec.Command("kind", "get", "clusters")
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return false
	}

	return strings.Contains(buf.String(), *cluster)
}

func preloadImages() {
	if !isKind() {
		return
	}

	for _, v := range preloadImagesList {
		loadImage(v.name, v.image, v.version)
	}
}

func loadImage(name, image, version string) {
	if version == "" {
		version = "latest"
	}

	img := image + ":" + version
	sayn("==> preloading image ", img)
	{
		// check if the image exists
		cmd := exec.Command("docker", "image", "inspect", img)
		if cmd.Run() != nil {
			// pull the image
			cmd = exec.Command("docker", "pull", img)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
		}
	}
	{
		cmd := exec.Command("kind", "load", "docker-image", img, "--name", *cluster)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

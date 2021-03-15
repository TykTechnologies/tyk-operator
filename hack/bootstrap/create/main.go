package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Config configuration for booting operator environment
type Config struct {
	WorkDir  string
	Tyk      Tyk
	Operator Operator
}

func (c *Config) bind(mode string) {
	c.Tyk.bind(mode)
	c.Operator.bind(mode)

	// set working directory
	wd, err := os.Getwd()
	if err != nil {
		exit(err)
	}
	c.WorkDir = filepath.Join(wd, "ci")
	env(&c.WorkDir, "TYK_OPERATOR_WORK_DIR")
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

func (t *Tyk) oss() {
	t.Mode = "oss"
	t.URL = "http://tyk.tykce-control-plane.svc.cluster.local:8001"
	t.Auth = "foo"
	t.Org = "myorg"
	t.Namespace = "tykce-control-plane"
}

func (t *Tyk) helm() {
	if t.Charts != "" {
		if t.Mode == "pro" {
			t.URL = "http://dashboard-svc-ci-tyk-pro.tykpro-control-plane.svc.cluster.local:3000"
		} else {
			t.URL = "http://gateway-svc-ci-tyk-headless.tykce-control-plane.svc.cluster.local:8000"
		}
	}
}

func (t *Tyk) pro() {
	t.Mode = "pro"
	t.URL = "http://dashboard.tykpro-control-plane.svc.cluster.local:3000"
	t.Namespace = "tykpro-control-plane"
	t.AdminSecret = "54321"
}

func (t *Tyk) bind(mode string) {
	if mode == "pro" {
		t.pro()
	} else {
		t.oss()
	}
	env(&t.Mode, "TYK_MODE")
	env(&t.URL, "TYK_URL")
	env(&t.Org, "TYK_ORG")
	env(&t.Auth, "TYK_AUTH")
	env(&t.Namespace, "TYK_NAMESPACE")
	env(&t.AdminSecret, "TYK_ADMIN_SECRET")
	env(&t.Charts, "TYK_HELM_CHARTS")
	env(&t.License, "TYK_DB_LICENSEKEY")
	if t.Charts != "" {
		// the control name is different for helm charts
		t.URL = "http://gateway-svc-tyk-headless.tykce-control-plane.svc.cluster.local:8000"
	}
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
	o.CertManager = "https://github.com/jetstack/cert-manager/releases/download/v1.0.4/cert-manager.yaml"
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

var mode = flag.String("mode", os.Getenv("TYK_MODE"), "ce for community and pro for pro")
var debug = flag.Bool("debug", false, "prints lots of details")

var charts = map[string]string{
	"oss": filepath.Join("charts", "tyk-headless"),
	"pro": filepath.Join("charts", "tyk-pro"),
}

func chartDir() string {
	return charts[config.Tyk.Mode]
}

var rep = map[string]string{
	"oss": "tyk-ce",
	"pro": "tyk-pro",
}

func deployDir() string {
	return rep[config.Tyk.Mode]
}

func main() {
	flag.Parse()
	config.bind(*mode)
	submodule()
	ns()
	common()
	if config.Tyk.Charts != "" {
		// when we have this provided we are installing the operator using official
		// helm charts
		helm()
		return
	}
	pro(dash)
	ce(community)
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
func bootsrapDash() {
	say("Bootstrapping dashboard ...")
	var buf bytes.Buffer
	exit(kf(func(c *exec.Cmd) {
		c.Stdout = &buf
		c.Stderr = os.Stderr
	},
		"exec", "-t", "-n", config.Tyk.Namespace,
		"svc/dashboard", "--", "./tyk-analytics", "bootstrap",
		"--conf", "/etc/tyk-dashboard/dash.json",
	))
	exit(ioutil.WriteFile("bootstrapped", buf.Bytes(), 0600))
	a := buf.Bytes()
	{
		u := "USER AUTHENTICATION CODE:"
		i := bytes.Index(a, []byte(u))
		n := bytes.TrimLeft(a[i+len(u):], " ")
		n = n[:bytes.Index(n, []byte("\n"))]
		config.Tyk.Auth = string(n)
	}
	{
		u := "ORG ID:"
		i := bytes.Index(a, []byte(u))
		n := bytes.TrimLeft(a[i+len(u):], " ")
		n = n[:bytes.Index(n, []byte("\n"))]
		config.Tyk.Org = string(n)
	}
	ok()
}

func pro(fn func()) {
	if config.Tyk.Mode == "pro" {
		fn()
	}
}

func ce(fn func()) {
	if config.Tyk.Mode == "oss" {
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
func ns() {
	say("Creating namespaces ...")
	if !hasNS(config.Tyk.Namespace) {
		exit(kl("create", "namespace", config.Tyk.Namespace))
	}
	if !hasNS(config.Operator.Namespace) {
		exit(kl("create", "namespace", config.Operator.Namespace))
	}
	ok()
}

func createDashSecret() {
	say("Creating dashboard secrets ...")
	var buf bytes.Buffer
	exit(kf(func(c *exec.Cmd) {
		c.Stdout = &buf
		c.Stderr = os.Stderr
	},
		"create", "secret", "-n", config.Tyk.Namespace,
		"generic", "dashboard", "--dry-run=client",
		"--from-literal", "license="+config.Tyk.License,
		"--from-literal", "adminSecret="+config.Tyk.AdminSecret, "-o", "yaml",
	))
	exit(kf(func(c *exec.Cmd) {
		c.Stdin = &buf
		c.Stderr = os.Stderr
	},
		"apply", "-f", "-",
	))
	ok()
}

func createSecret() {
	say("Creating Secret ... ")
	if !hasOperatorSecret() {
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
	if !hasRedis() {
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
	if !hasMongo() {
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

func helm() {
	createRedis()
	say("Installing helm chart ...")
	if !hasChart() {
		c := filepath.Join(config.Tyk.Charts, chartDir())
		f := filepath.Join(config.WorkDir, "helm", chartDir(), "values.yaml")
		cmd := exec.Command("helm", "install", config.Tyk.Mode,
			"-f", f,
			c,
			"-n", config.Tyk.Namespace,
			"--wait",
		)
		cmd.Dir = config.WorkDir
		cmd.Stderr = os.Stderr
		if *debug {
			cmd.Stdout = os.Stdout
			fmt.Println(cmd.Args)
		}
		exit(cmd.Run())
	}
	ok()
}

func hasChart() bool {
	cmd := exec.Command("helm", "list",
		"-n", config.Tyk.Namespace,
	)
	cmd.Dir = config.WorkDir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	exit(cmd.Run())
	return strings.Contains(buf.String(), config.Tyk.Mode)
}

func hasNS(name string) bool {
	return k("get", "ns", name) == nil
}

func hasSecret(name string) bool {
	return k("get", "secret", "-n", config.Tyk.Namespace, name) == nil
}

func hasOperatorSecret() bool {
	return k("get", "secret", "-n", config.Operator.Namespace, config.Operator.SecretName) == nil
}
func hasRedis() bool {
	return k("get", "deployment/redis", "-n", config.Tyk.Namespace) == nil
}

func hasMongo() bool {
	return k("get", "deployment/mongo", "-n", config.Tyk.Namespace) == nil
}

func hasGateway() bool {
	return k("get", "deployment/tyk", "-n", config.Tyk.Namespace) == nil
}

func hasDash() bool {
	return k("get", "deployment/dashboard", "-n", config.Tyk.Namespace) == nil
}

func hasHTTPBIN() bool {
	return k("get", "deployment/httpbin") == nil
}
func hasGRPCPlugin() bool {
	return k("get", "deployment/grpc-plugin", "-n", config.Tyk.Namespace) == nil
}

func hasCertManager() bool {
	return k("get", "deployment/cert-manager", "-n", config.Operator.CertManagerNamespace) == nil
}

func hasConfigMap(name string) bool {
	return k("get", "configmap", name, "-n", config.Tyk.Namespace) == nil
}

func createCertManager() {
	say("Installing cert-manager ...")
	if !hasCertManager() {
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

func createConfigMaps() {
	{
		say("Creating dashboard configmap ...")
		name := "dash-conf"
		if !hasConfigMap(name) {
			var buf bytes.Buffer
			exit(kf(func(c *exec.Cmd) {
				c.Stdout = &buf
				c.Stderr = os.Stderr
			},
				"create", "configmap",
				"-n", config.Tyk.Namespace,
				name, "--dry-run=client",
				"--from-file", filepath.Join(config.WorkDir, deployDir(), "dashboard", "confs", "dash.json"),
				"-o", "yaml",
			))
			exit(kf(func(c *exec.Cmd) {
				c.Stdin = &buf
				c.Stderr = os.Stderr
			},
				"apply", "-f", "-",
			))
		}
		ok()
	}
	{
		say("Creating tyk configmap ...")
		name := "tyk-conf"
		if !hasConfigMap(name) {
			var buf bytes.Buffer
			exit(kf(func(c *exec.Cmd) {
				c.Stdout = &buf
				c.Stderr = os.Stderr
			},
				"create", "configmap",
				"-n", config.Tyk.Namespace,
				name, "--dry-run=client",
				"--from-file", filepath.Join(config.WorkDir, deployDir(), "gateway", "confs", "tyk.json"),
				"-o", "yaml",
			))
			exit(kf(func(c *exec.Cmd) {
				c.Stdin = &buf
				c.Stderr = os.Stderr
			},
				"apply", "-f", "-",
			))
		}
		ok()
	}
}

func dash() {
	createMongo()
	createConfigMaps()
	createDashSecret()
	deployDash()
	deployGateway()
	bootsrapDash()
}

func common() {
	deployHTTPBIN()
	deployGRPCPlugin()
	createRedis()
}

func community() {
	communityConfigMap()
	deployGateway()
}

func communityConfigMap() {
	say("creating config maps ...")
	if !hasConfigMap("tyk-conf") {
		exit(kl("create", "configmap",
			"-n", config.Tyk.Namespace,
			"tyk-conf",
			"--from-file", filepath.Join(config.WorkDir, deployDir(), "gateway", "confs", "tyk.json"),
		))
	}
	ok()
}

func operator() {
	createCertManager()
	createSecret()
}

func deployDash() {
	say("Deploying dashboard ...")
	if !hasDash() {
		exit(k("apply", "-n", config.Tyk.Namespace, "-f", filepath.Join(config.WorkDir, deployDir(), "dashboard")))
		ok()
		say("Waiting for dashboard to be ready ...")
		exit(k(
			"rollout", "status", "deployment/dashboard", "-n", config.Tyk.Namespace,
		))
	}
	ok()
}

func deployGateway() {
	say("Deploying gateway ...")
	if !hasGateway() {
		exit(k("apply", "-n", config.Tyk.Namespace, "-f", filepath.Join(config.WorkDir, deployDir(), "gateway")))
		ok()
		say("Waiting for gateway to be ready ...")
		exit(k(
			"rollout", "status", "deployment/tyk", "-n", config.Tyk.Namespace,
		))
	}
	ok()
}

func deployHTTPBIN() {
	say("Deploying httpbin ...")
	if !hasHTTPBIN() {
		exit(k("apply", "-f", filepath.Join(config.WorkDir, "upstreams")))
		ok()
		say("Waiting for httpbin to be ready ...")
		exit(kl(
			"rollout", "status", "deployment/httpbin",
		))
	}
	ok()
}

func deployGRPCPlugin() {
	say("Deploying grpc-plugin ...")
	if !hasGRPCPlugin() {
		exit(k("apply", "-f", filepath.Join(config.WorkDir, "grpc-plugin"), "-n", config.Tyk.Namespace))
		ok()
		say("Waiting for grpc-plugin to be ready ...")
		exit(kl(
			"rollout", "status", "deployment/grpc-plugin", "-n", config.Tyk.Namespace,
		))
	}
	ok()
}

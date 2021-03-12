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

var auth = flag.String("auth", os.Getenv("TYK_AUTH"), "Secret")
var org = flag.String("org", os.Getenv("TYK_ORG"), "organizarion id")
var mode = flag.String("mode", os.Getenv("TYK_MODE"), "ce for community and pro for pro")
var target = flag.String("url", os.Getenv("TYK_URL"), "ce for community and pro for pro")
var secret = flag.String("secret", "tyk-operator-conf", "ce for community and pro for pro")
var namespace = flag.String("ns", "", "namespace")
var operatorNS = flag.String("op-ns", "tyk-operator-system", "namespace")
var debug = flag.Bool("debug", false, "prints lots of details")
var certManagerRelease = flag.String("certmanager-release", "https://github.com/jetstack/cert-manager/releases/download/v1.0.4/cert-manager.yaml", "The file for cert manager")
var certManagerNamespace = flag.String("cert-manager-ns", "cert-manager", "The namespace to install cert manager")
var adminSecret = flag.String("admin-secret", "54321", "admin-secret")
var workdir string

var charts = map[string]string{
	"oss": filepath.Join("charts", "tyk-headless"),
	"pro": filepath.Join("charts"),
}

var repo = map[string]string{
	"oss": "tyk-ce",
	"pro": "tyk-pro",
}

func init() {
	wd, err := os.Getwd()
	if err != nil {
		exit(err)
	}
	workdir = filepath.Join(wd, "ci")
}

func main() {
	flag.Parse()
	bootsrap()
	if flag.Arg(0) == "down" {
		down()
		return
	}
	createNS()
	pro(dash)
	ce(community)
	operator()
}

func bootsrap() {
	if *mode == "pro" {
		// remove bootstrapped file so we can bootsrap a new one
		os.Remove("bootstrapped")
		if *target == "" {
			*target = "http://dashboard.tykpro-control-plane.svc.cluster.local:3000"
		}
		if *namespace == "" {
			*namespace = "tykpro-control-plane"
		}
	} else {
		// we are running community edition
		if *target == "" {
			*target = "http://tyk.tykce-control-plane.svc.cluster.local:8001"
		}
		if *auth == "" {
			*auth = "foo"
		}
		if *org == "" {
			*org = "myorg"
		}
		*mode = "oss"
		if *namespace == "" {
			*namespace = "tykce-control-plane"
		}
	}
}

func bootsrapDash() {
	say("Bootstrapping dashboard ...")
	_, err := os.Stat("bootstrapped")
	if err != nil {
		var buf bytes.Buffer
		exit(kf(func(c *exec.Cmd) {
			c.Stdout = &buf
			c.Stderr = os.Stderr
		},
			"exec", "-t", "-n", *namespace,
			"svc/dashboard", "--", "./tyk-analytics", "bootstrap",
			"--conf", "/etc/tyk-dashboard/dash.json",
		))
		exit(ioutil.WriteFile("bootstrapped", buf.Bytes(), 0600))
	}
	a, err := ioutil.ReadFile("bootstrapped")
	if err != nil {
		exit(err)
	}
	{
		u := "USER AUTHENTICATION CODE:"
		i := bytes.Index(a, []byte(u))
		n := bytes.TrimLeft(a[i+len(u):], " ")
		n = n[:bytes.Index(n, []byte("\n"))]
		*auth = string(n)
	}
	{
		u := "ORG ID:"
		i := bytes.Index(a, []byte(u))
		n := bytes.TrimLeft(a[i+len(u):], " ")
		n = n[:bytes.Index(n, []byte("\n"))]
		*org = string(n)
	}
	ok()
	fmt.Println("ORG ", *org)
	fmt.Println("AUTH ", *auth)
}

// down removes  installed resources
func down() {
	say("Uninstalling helm chart ...")
	if hasChart() {
		exit(exec.Command("helm", "uninstall", *mode, "-n", *namespace).Run())
	}
	ok()
	say("Deleting secret ...")
	if hasOperatorSecret() {
		exit(k("delete", "secret", *secret, "-n", *operatorNS))
	}
	ok()
	say("Deleting redis ...")
	if hasRedis() {
		f := filepath.Join(workdir, repo[*mode], "redis")
		exit(k("delete", "-f", f, "-n", *namespace))
	}
	ok()
	pro(func() {
		say("Deleting mongo ...")
		if hasRedis() {
			f := filepath.Join(workdir, repo[*mode], "mongo")
			exit(k("delete", "-f", f, "-n", *namespace))
		}
		ok()
	})
	pro(func() {
		say("Deleting configmaps ...")
		if hasConfigMap("dash-conf") {
			exit(k("delete", "configmap", "dash-conf", "-n", *namespace))
		}
		if hasConfigMap("tyk-conf") {
			exit(k("delete", "configmap", "tyk-conf", "-n", *namespace))
		}
		ok()
	})
	say("Deleting cert manager ...")
	if hasCertManager() {
		exit(k("delete", "-f", *certManagerRelease))
	}
	ok()
}

func pro(fn func()) {
	if *mode == "pro" {
		fn()
	}
}

func ce(fn func()) {
	if *mode == "oss" {
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

func kf(fn func(*exec.Cmd), args ...string) error {
	cmd := exec.Command("kubectl", args...)
	cmd.Dir = workdir
	fn(cmd)
	if *debug {
		fmt.Println("==>", cmd.Args)
	}
	return cmd.Run()
}

func createNS() {
	say("Creating namespace ...")
	if !hasNS(*namespace) {
		exit(k("create", "namespace", *namespace))
	}
	if !hasNS(*operatorNS) {
		exit(k("create", "namespace", *operatorNS))
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
		"create", "secret", "-n", *namespace,
		"generic", "dashboard", "--dry-run=client",
		"--from-literal", "license="+os.Getenv("TYK_DB_LICENSEKEY"),
		"--from-literal", "adminSecret="+*adminSecret, "-o", "yaml",
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
		exit(k("create", "secret",
			"-n", *operatorNS,
			"generic", *secret,
			"--from-literal", fmt.Sprintf("TYK_AUTH=%s", *auth),
			"--from-literal", fmt.Sprintf("TYK_ORG=%s", *org),
			"--from-literal", fmt.Sprintf("TYK_MODE=%s", *mode),
			"--from-literal", fmt.Sprintf("TYK_URL=%s", *target),
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
		f := filepath.Join(workdir, repo[*mode], "redis")
		exit(k(
			"apply", "-f", f,
			"-n", *namespace,
		))
		ok()
		say("Waiting for redis to be ready ...")
		exit(k(
			"rollout", "status", "deployment/redis", "-n", *namespace,
		))
	}
	ok()
}

func createMongo() {
	say("Creating Mongo ....")
	if !hasMongo() {
		f := filepath.Join(workdir, repo[*mode], "mongo")
		exit(k(
			"apply", "-f", f,
			"-n", *namespace,
		))
		ok()
		say("Waiting for mongo to be ready ...")
		exit(k(
			"rollout", "status", "deployment/mongo", "-n", *namespace,
		))
	}
	ok()
}

func createHelm() {
	say("Installing helm chart ...")
	if !hasChart() {
		c := filepath.Join(workdir, charts[*mode])
		f := filepath.Join(workdir, repo[*mode], "values.yaml")
		cmd := exec.Command("helm", "install", *mode,
			"-f", f,
			c,
			"-n", *namespace,
			"--wait",
		)
		fmt.Println(cmd.Args)
		cmd.Dir = workdir
		cmd.Stderr = os.Stderr
		exit(cmd.Run())
	}
	ok()
}

func hasChart() bool {
	cmd := exec.Command("helm", "list",
		"-n", *namespace,
	)
	cmd.Dir = workdir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	exit(cmd.Run())
	return strings.Contains(buf.String(), *mode)
}

func hasNS(name string) bool {
	return k("get", "ns", name) == nil
}

func hasSecret(name string) bool {
	return k("get", "secret", "-n", *namespace, name) == nil
}

func hasOperatorSecret() bool {
	return k("get", "secret", "-n", *operatorNS, *secret) == nil
}
func hasRedis() bool {
	return k("get", "deployment/redis", "-n", *namespace) == nil
}

func hasMongo() bool {
	return k("get", "deployment/mongo", "-n", *namespace) == nil
}

func hasGateway() bool {
	return k("get", "deployment/tyk", "-n", *namespace) == nil
}

func hasDash() bool {
	return k("get", "deployment/dashboard", "-n", *namespace) == nil
}

func hasHTTPBIN() bool {
	return k("get", "deployment/httpbin") == nil
}

func hasCertManager() bool {
	return k("get", "deployment/cert-manager", "-n", *certManagerNamespace) == nil
}

func hasConfigMap(name string) bool {
	return k("get", "configmap", name, "-n", *namespace) == nil
}

func createCertManager() {
	say("Installing cert-manager ...")
	if !hasCertManager() {
		exit(k(
			"apply",
			"--validate=false",
			"-f", *certManagerRelease,
		))
		ok()
		say("Waiting for cert-manager to be ready ...")
		exit(k(
			"rollout", "status", "deployment/cert-manager", "-n", *certManagerNamespace,
		))
		exit(k(
			"rollout", "status", "deployment/cert-manager-cainjector", "-n", *certManagerNamespace,
		))
		exit(k(
			"rollout", "status", "deployment/cert-manager-webhook", "-n", *certManagerNamespace,
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
				"-n", *namespace,
				name, "--dry-run=client",
				"--from-file", filepath.Join(workdir, repo[*mode], "dashboard", "confs", "dash.json"),
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
				"-n", *namespace,
				name, "--dry-run=client",
				"--from-file", filepath.Join(workdir, repo[*mode], "gateway", "confs", "tyk.json"),
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
	createRedis()
	createMongo()
	createConfigMaps()
	createDashSecret()
	deployDash()
	deployGateway()
	deployHTTPBIN()
	bootsrapDash()
}

func community() {
	createRedis()
	createHelm()
}

func operator() {
	createCertManager()
	createSecret()
}

func deployDash() {
	say("Deploying dashboard ...")
	if !hasDash() {
		exit(k("apply", "-n", *namespace, "-f", filepath.Join(workdir, repo[*mode], "dashboard")))
		ok()
		say("Waiting for dashboard to be ready ...")
		exit(k(
			"rollout", "status", "deployment/dashboard", "-n", *namespace,
		))
	}
	ok()
}

func deployGateway() {
	say("Deploying gateway ...")
	if !hasGateway() {
		exit(k("apply", "-n", *namespace, "-f", filepath.Join(workdir, repo[*mode], "gateway")))
		ok()
		say("Waiting for gateway to be ready ...")
		exit(k(
			"rollout", "status", "deployment/tyk", "-n", *namespace,
		))
	}
	ok()
}

func deployHTTPBIN() {
	say("Deploying httpbin ...")
	if !hasHTTPBIN() {
		exit(k("apply", "-f", filepath.Join(workdir, "upstreams")))
		ok()
		say("Waiting for httpbin to be ready ...")
		exit(k(
			"rollout", "status", "deployment/httpbin",
		))
	}
	ok()
}

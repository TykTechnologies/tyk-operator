package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
)

var image = flag.String("image", "", "docker image for the operator")
var cluster = flag.String("cluster", "", "cluster name")

func main() {
	flag.Parse()

	if IsKind() {
		cmd := exec.Command("kind", "load", "docker-image", *image, "--name", *cluster)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout

		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}

	if IsMinikube() {
		// Push image to minikube
		// source https://minikube.sigs.k8s.io/docs/handbook/pushing/#2-push-images-using-cache-command
		cmd := exec.Command("minikube", "cache", "add", *image)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}

		cmd = exec.Command("minikube", "cache", "reload")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

// IsKind returns true if we have kind cluster
func IsKind() bool {
	cmd := exec.Command("kind", "get", "clusters")

	var buf bytes.Buffer
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return false
	}

	return strings.Contains(buf.String(), *cluster)
}

// IsMinikube returns true if we are running minikube
func IsMinikube() bool {
	cmd := exec.Command("minikube", "status")

	var buf bytes.Buffer
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return false
	}
	// We count Running twice because we need both minikube and cluster to be
	// running.
	return strings.Count(buf.String(), "Running") == 2
}

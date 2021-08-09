package integration

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

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
	cmd := exec.Command("kind", "get", "kubeconfig")
	cmd.Stdout = o
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func setupKind() (string, error) {
	if !isKind() {
		return "", errors.New("Missing kind cluster")
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

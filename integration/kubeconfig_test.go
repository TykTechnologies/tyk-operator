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
	cmd := exec.Command("kind", "get", "clusters")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return false
	}
	return strings.Contains(buf.String(), "kind")
}

func kubeConf(o io.Writer) error {
	cmd := exec.Command("kind", "get", "kubeconfig")
	cmd.Stdout = o
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
		return "", os.RemoveAll(f.Name())
	}
	defer f.Close()

	return f.Name(), nil
}

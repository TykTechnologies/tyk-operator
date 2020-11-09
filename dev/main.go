package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// clusterName   the name of the development cluster.
const clusterName = "kind"

// registry is the name of the local registry container.
const registry = "kind-registry"

// containerPath the directory in the container that is mounted to the local
// node directory
const containerPath = "/data"

// registryPort the port exposed by the registry container.
const registryPort = 5000

func main() {
	log.SetFlags(log.Lshortfile)
	o := opts{}
	flag.StringVar(&o.RegistryContainer.Name, "r", registry, "The name of the registry container")
	flag.IntVar(&o.RegistryContainer.Port, "p", registryPort, "The port on which registry is exposed")
	flag.StringVar(&o.Cluster.Name, "c", clusterName, "The name of the developent cluster")
	flag.IntVar(&o.Cluster.Nodes, "n", 4, "The number of nodes in a kind cluster")
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	wd = filepath.Join(wd, "tmp", "kind")
	flag.StringVar(&o.WorkDir, "d", wd, "The directory for mounting nodes ")
	flag.Parse()
	switch flag.Arg(0) {
	case "up":
		err := up(context.Background(), o)
		if err != nil {
			log.Fatalf("Failed to boot with :%v", err)
		}
	case "down":
		err := down(o.Cluster.Name)
		if err != nil {
			log.Fatalf("Failed to tear down cluster with :%v", err)
		}
	case "cert":
		err := certManager()
		if err != nil {
			log.Fatalf("Failed to install cert manager with :%v", err)
		}
	}
}

type opts struct {
	RegistryContainer struct {
		Name string
		Port int
	}
	Cluster struct {
		Name  string
		Nodes int
	}
	WorkDir string
}

func (o opts) patch() string {
	return fmt.Sprintf(patchTpl, o.RegistryContainer.Port,
		o.RegistryContainer.Name, o.RegistryContainer.Port)
}

func down(name string) error {
	cmd := exec.Command("kind", "delete", "cluster", "--name", name)
	command(cmd)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// up setup directory to be mounted on kind nodes and returns cluster
// configuration.
//
// We create a 4 node cluster named node-{0,1,2,3} we mount work/node-{0,1,2,3}
// to respective node.
//
// node-0 is the control plane node
func up(ctx context.Context, o opts) error {
	_, err := exec.LookPath("kind")
	if err != nil {
		return errors.New("kind binary not in $PATH")
	}
	_, err = exec.LookPath("docker")
	if err != nil {
		return errors.New("docker binary not in $PATH")
	}
	ok, err := hasCluster(ctx, o.Cluster.Name)
	if err != nil {
		return err
	}
	if ok {
		log.Printf("cluster %q already exists\n", o.Cluster.Name)
		return nil
	}
	ok, err = configDocker(ctx)
	if err != nil {
		return err
	}
	if !ok {
		if err := createRegistry(ctx); err != nil {
			return err
		}
	}
	n, err := nodes(o.WorkDir, o.Cluster.Nodes)
	if err != nil {
		return err
	}

	conf := &v1alpha4.Cluster{
		TypeMeta: v1alpha4.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
		Nodes:                   n,
		ContainerdConfigPatches: []string{strings.TrimSpace(o.patch())},
	}
	b, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	err = startCluster(ctx, o.Cluster.Name, b)
	if err != nil {
		return err
	}
	ok, err = hasNetwork(ctx)
	if err != nil {
		return err
	}
	if !ok {
		if err := connect(ctx); err != nil {
			return err
		}
	}
	err = annotate(ctx)
	if err != nil {
		return err
	}
	return certManager()
}

// const patch = `
// [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
// 	endpoint = ["http://kind-registry:5000"]
// `
const patchTpl = `
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:%d"]
	endpoint = ["http://%s:%d"]
`

// nodes return configuration for nodes in a kind cluster. n is the number of
// nodes. The nodes are labelled node-{0...n-1}. This will make sure the node
// directory is created if it doesn't exist yet.
func nodes(workDir string, n int) (nodes []v1alpha4.Node, err error) {
	if n == 0 {
		n = 4
	}
	for k := 0; k < 4; k++ {
		name := fmt.Sprintf("node-%d", k)
		path := filepath.Join(workDir, name)
		if err := CheckDir(path); err != nil {
			return nil, err
		}
		role := v1alpha4.WorkerRole
		if k == 0 {
			role = v1alpha4.ControlPlaneRole
		}
		nodes = append(nodes, v1alpha4.Node{
			Role: role,
			ExtraMounts: []v1alpha4.Mount{
				{
					HostPath:      path,
					ContainerPath: containerPath,
				},
			},
		})
	}
	return
}

func configDocker(ctx context.Context) (bool, error) {
	var buf bytes.Buffer
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", registry)
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	if buf.Len() == 0 {
		return false, errors.New("Failed to get registry container status")
	}
	return strings.TrimSpace(buf.String()) == "true", nil
}

func createRegistry(ctx context.Context) error {
	cmd := exec.Command("docker", "run", "-d", "--restart=always",
		"-p", fmt.Sprintf("%d:5000", registryPort), "--name",
		registry, "registry:2",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	command(cmd)
	return cmd.Run()
}

func command(cmd *exec.Cmd) {
	a := strings.Join(cmd.Args, " ")
	log.Println("Exec " + a)
}

func hasNetwork(ctx context.Context) (bool, error) {
	cmd := exec.Command("docker", "network", "ls", "-f",
		fmt.Sprintf("name=%s", clusterName), "--format", "{{.Name}}",
	)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	command(cmd)
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	if buf.Len() == 0 {
		return false, nil
	}
	return strings.Contains(buf.String(), clusterName), nil
}

func connect(ctx context.Context) error {
	cmd := exec.Command("docker", "network", "connect", clusterName, registry)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	command(cmd)
	return cmd.Run()
}

func startCluster(ctx context.Context, name string, conf []byte) error {
	cmd := exec.Command("kind", "create", "cluster", "--config", "-")
	if name != "" {
		cmd.Args = append(cmd.Args, "--name", name)
	}
	cmd.Stdin = bytes.NewReader(conf)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	command(cmd)
	return cmd.Run()
}

func listClusters(ctx context.Context) ([]string, error) {
	cmd := exec.Command("kind", "get", "clusters")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	command(cmd)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(&buf)
	s.Split(bufio.ScanLines)
	var o []string
	for s.Scan() {
		o = append(o, strings.TrimSpace(s.Text()))
	}
	return o, nil
}

func hasCluster(ctx context.Context, name string) (bool, error) {
	ls, err := listClusters(ctx)
	if err != nil {
		return false, err
	}
	for _, v := range ls {
		if v == name {
			return true, nil
		}
	}
	return false, nil
}

func listNodes(ctx context.Context) ([]string, error) {
	cmd := exec.Command("kind", "get", "nodes")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	command(cmd)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(&buf)
	s.Split(bufio.ScanLines)
	var o []string
	for s.Scan() {
		o = append(o, strings.TrimSpace(s.Text()))
	}
	return o, nil
}
func annotate(ctx context.Context) error {
	nodes, err := listNodes(ctx)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if err := annotateCMD(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

func annotateCMD(ctx context.Context, node string) error {
	cmd := exec.Command("kubectl", "annotate", "node", node,
		fmt.Sprintf("kind.x-k8s.io/registry=localhost:%d", registryPort),
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	command(cmd)
	return cmd.Run()
}

const perm = 0600

// CheckDir validates that with is directory. If with does not exist an attemp
// to create one will be done. A temporary file will be created in with to make sure it is writable.
func CheckDir(with string) error {
	stat, err := os.Stat(with)
	if err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("%v is not a directory", with)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	err = os.MkdirAll(with, os.ModePerm)
	if err != nil {
		return err
	}
	check := filepath.Join(with, "check.txt")
	err = ioutil.WriteFile(check, []byte("check"), perm)
	if err != nil {
		return err
	}
	return os.Remove(check)
}

func certManager() error {
	cmd := exec.Command("kubectl", "apply", "--validate=false", "-f",
		"https://github.com/jetstack/cert-manager/releases/download/v1.0.4/cert-manager.yaml",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	command(cmd)
	return cmd.Run()
}

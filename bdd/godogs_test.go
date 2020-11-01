package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"
)

const (
	namespace = "bdd"
)

type store struct {
	gatewayNamespace string
	responseCode     int
	cleanupK8s       []string
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	s := store{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		app := "kubectl"

		s.gatewayNamespace = fmt.Sprintf("tyk%s-control-plane", os.Getenv("TYK_MODE"))

		cmd := exec.Command(app, "create", "ns", namespace)
		output, err := cmd.Output()
		if err != nil {
			panic(err)
		}
		if !strings.Contains(string(output), fmt.Sprintf("namespace/%s created", namespace)) {
			panic(string(output))
		}
	})

	ctx.AfterScenario(func(sc *godog.Scenario, err error) {
		app := "kubectl"

		cmd := exec.Command(app, "delete", "ns", namespace)
		output, err := cmd.Output()
		if err != nil {
			panic(err)
		}
		if !strings.Contains(string(output), fmt.Sprintf(`namespace "%s" deleted`, namespace)) {
			panic(string(output))
		}
	})

	ctx.Step(`^there is a (\S+) resource$`, s.thereIsAResource)
	ctx.Step(`^i create a (\S+) resource$`, s.iCreateAResource)
	ctx.Step(`^i update a (\S+) resource$`, s.iUpdateAResource)
	ctx.Step(`^i delete a (\S+) resource$`, s.iRequestEndpoint)
	ctx.Step(`^i request (\S+) endpoint$`, s.iRequestEndpoint)
	ctx.Step(`^there should be a (\d+) http response code$`, s.thereShouldBeHttpResponseCode)
}

// waitForServices tests and waits on the availability of a TCP host and port
func waitForServices(services []string, timeOut time.Duration) error {
	var depChan = make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(len(services))
	go func() {
		for _, s := range services {
			go func(s string) {
				defer wg.Done()
				for {
					_, err := net.Dial("tcp", s)
					if err == nil {
						return
					}
					time.Sleep(1 * time.Second)
				}
			}(s)
		}
		wg.Wait()
		close(depChan)
	}()

	select {
	case <-depChan: // services are ready
		return nil
	case <-time.After(timeOut):
		return fmt.Errorf("services aren't ready in %s", timeOut)
	}
}

func (s *store) iRequestEndpoint(path string) error {
	cmd := exec.Command("kubectl", "port-forward", "-n", s.gatewayNamespace, "svc/gw", "8000:8000")
	err := cmd.Start()
	if err != nil {
		return err
	}
	defer cmd.Process.Kill()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			conn, err := net.DialTimeout("tcp", "127.0.0.1:8000", time.Second*3)
			if err != nil {
				time.Sleep(time.Millisecond * 500)
				continue
			}
			if conn != nil {
				conn.Close()
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()

	res, err := http.Get(fmt.Sprintf("http://localhost:8000%s", path))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	s.responseCode = res.StatusCode

	return nil
}

func (s *store) thereIsAResource(fileName string) error {
	return s.kubectlFile("apply", fileName, " created", time.Second*10)
}

func (s *store) iCreateAResource(fileName string) error {
	return s.kubectlFile("apply", fileName, " created", time.Second*10)
}

func (s *store) iUpdateAResource(fileName string) error {
	return s.kubectlFile("apply", fileName, " configured", time.Second*10)
}

func (s *store) iDeleteAResource(fileName string) error {
	return s.kubectlFile("delete", fileName, " deleted", time.Second*10)
}

func (s *store) kubectlFile(action string, fileName string, expected string, timeout time.Duration) error {
	app := "kubectl"

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, app, action, "-f", fileName, "-n", namespace)
	output, err := cmd.Output()

	if err != nil {
		return err
	}

	if !strings.Contains(string(output), expected) {
		return fmt.Errorf("unexpected output (%s)", string(output))
	}

	return nil
}

func (s *store) thereShouldBeHttpResponseCode(expectedCode int) error {
	if expectedCode != s.responseCode {
		return fmt.Errorf("expected http status %d, got http %d", expectedCode, s.responseCode)
	}

	return nil
}

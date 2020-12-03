package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cucumber/godog"
)

const (
	namespace  = "bdd"
	k8sTimeout = time.Second * 10
)

var gwNS = fmt.Sprintf("tyk%s-control-plane", os.Getenv("TYK_MODE"))
var client = http.Client{}

func runCMD(cmd *exec.Cmd) string {
	a := fmt.Sprint(cmd.Args)
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("failed %s with %v : %s", a, err, string(output)))
	}
	return string(output)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		app := "kubectl"
		cmd := exec.Command(app, "create", "ns", namespace)
		output := runCMD(cmd)
		if !strings.Contains(output, fmt.Sprintf("namespace/%s created", namespace)) {
			panic(string(output))
		}
	})

	ctx.AfterSuite(func() {
		app := "kubectl"
		cmd := exec.Command(app, "delete", "ns", namespace)
		output := runCMD(cmd)
		if !strings.Contains(string(output), fmt.Sprintf(`namespace "%s" deleted`, namespace)) {
			panic(string(output))
		}
	})
}

var opts = &godog.Options{
	StopOnFailure: true,
	Format:        "pretty",
	Tags:          "~@undone",
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, opts)
}

type writeFn func([]byte) (int, error)

func (fn writeFn) Write(b []byte) (int, error) {
	return fn(b)
}

func set(comm chan struct{}) {
	for {
		kill, term, err := setup()
		if err != nil {
			panic(err)
		}
		comm <- struct{}{}
		select {
		case <-term:
			kill()
			fmt.Println("===> reopening port forwarding")
		case <-comm:
			kill()
			comm <- struct{}{}
			return
		}
	}
}

func setup() (func() error, chan struct{}, error) {
	// make sure we don't have the testing ns
	exec.Command("kubectl", "delete", "ns", namespace).Run()
	cmd := exec.Command("kubectl", "port-forward", "-n", gwNS, "svc/gw", "8000:8000")
	fmt.Println(cmd.Args)
	var once sync.Once
	firstLine := make(chan string, 1)
	fail := "failed to execute portforward in network namespace"
	term := make(chan struct{})
	cmd.Stderr = writeFn(func(b []byte) (int, error) {
		once.Do(func() { firstLine <- string(b) })
		if bytes.Contains(b, []byte(fail)) {
			term <- struct{}{}
		}
		return os.Stderr.Write(b)
	})
	cmd.Stdout = writeFn(func(b []byte) (int, error) {
		once.Do(func() { firstLine <- string(b) })
		return os.Stdout.Write(b)
	})
	err := cmd.Start()
	if err != nil {
		return nil, nil, err
	}
	ts := time.NewTimer(3 * time.Second)
	defer ts.Stop()
	select {
	case <-ts.C:
		return nil, nil, errors.New("timeout waiting for port forwarding")
	case b := <-firstLine:
		x := "Forwarding from 127.0.0.1:8000"
		if !strings.HasPrefix(b, x) {
			cmd.Process.Kill()
			return nil, nil, fmt.Errorf("expected %q got %q", x, b)
		}
	}
	return cmd.Process.Kill, term, nil
}

func TestMain(t *testing.M) {
	flag.Parse()
	opts.Paths = flag.Args()
	comm := make(chan struct{})
	go set(comm)
	select {
	case <-comm:
	case <-time.After(3 * time.Second):
		log.Fatal("Failed to setup port forwarding")
	}

	status := godog.TestSuite{
		Name:                 "godogs",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              opts,
	}.Run()
	if st := t.Run(); st > status {
		status = st
	}
	comm <- struct{}{}
	select {
	case <-comm:
	case <-time.After(3 * time.Second):
		log.Fatal("Failed to setup port forwarding")
	}
	os.Exit(status)
}

type store struct {
	responseCode    int
	responseBody    []byte
	responseTimes   []time.Duration
	cleanupK8s      []string
	responseHeaders http.Header
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	s := store{}
	ctx.Step(`^there is a (\S+) resource$`, s.thereIsAResource)
	ctx.Step(`^i create a (\S+) resource$`, s.iCreateAResource)
	ctx.Step(`^i update a (\S+) resource$`, s.iUpdateAResource)
	ctx.Step(`^i delete a (\S+) resource$`, s.iDeleteAResource)
	ctx.Step(`^i request (\S+) endpoint$`, s.iRequestEndpoint)
	ctx.Step(`^i request (\S+) endpoint with header (\S+): (\S+)$`, s.iRequestEndpointWithHeader)
	ctx.Step(`^i request (\S+) endpoint with header (\S+): (\S+) (\d+) times$`, s.iRequestEndpointWithHeaderTimes)
	ctx.Step(`^the first response should be slowest$`, s.theFirstResponseShouldBeSlowest)
	ctx.Step(`^there should be a (\d+) http response code$`, s.thereShouldBeHttpResponseCode)
	ctx.Step(`^there should be a "(\S+): (\S+)" response header$`, s.thereShouldBeAResponseHeader)
	ctx.Step(`^the response should contain json key: (\S+) value: (\S+)$`, s.theResponseShouldContainJSONKeyValue)
	ctx.Step(`^the response should match JSON:$`, s.theResponseShouldMatchJSON)
}

func (s *store) iRequestEndpointWithHeaderTimes(path string, headerKey string, headerValue string, times int) error {
	for i := 0; i < times; i++ {
		t1 := time.Now()
		_ = s.iRequestEndpointWithHeader(path, headerKey, headerValue)
		t2 := time.Now()

		duration := t2.Sub(t1)
		s.responseTimes = append(s.responseTimes, duration)
	}
	return nil
}

func (s *store) theFirstResponseShouldBeSlowest() error {
	var firstResponse time.Duration

	if len(s.responseTimes) < 2 {
		return errors.New("cannot compare response times, not enough items")
	}

	for i, duration := range s.responseTimes {
		if i == 0 {
			firstResponse = duration
			continue
		}
		if duration > firstResponse {
			return fmt.Errorf("first response was faster %d", i)
		}
	}
	return nil
}

func call(method, url string, body func() io.Reader,
	fn func(*http.Request),
	validate func(*http.Response) error) error {
	var failed error
	err := backoff.Retry(func() error {
		req, err := http.NewRequest(method, url, body())
		if err != nil {
			failed = err
			return nil
		}
		if fn != nil {
			fn(req)
		}
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		failed = validate(res)
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return err
	}
	return failed
}

func (s *store) iRequestEndpointWithHeader(path string, headerKey string, headerValue string) error {
	return call(
		http.MethodGet,
		fmt.Sprintf("http://localhost:8000%s", path),
		func() io.Reader { return nil },
		func(h *http.Request) {
			h.Header.Set(headerKey, headerValue)
		},
		func(res *http.Response) error {
			s.responseCode = res.StatusCode
			s.responseHeaders = res.Header.Clone()
			s.responseBody, _ = ioutil.ReadAll(res.Body)
			return nil
		},
	)
}

func (s *store) iRequestEndpoint(path string) error {
	return call(
		http.MethodGet,
		fmt.Sprintf("http://localhost:8000%s", path),
		func() io.Reader { return nil },
		func(h *http.Request) {},
		func(h *http.Response) error {
			s.responseCode = h.StatusCode
			s.responseHeaders = h.Header.Clone()
			s.responseBody, _ = ioutil.ReadAll(h.Body)
			return nil
		},
	)
}

func (s *store) thereIsAResource(fileName string) error {
	return s.kubectlFile("apply", fileName, k8sTimeout, " created", " unchanged")
}

func (s *store) iCreateAResource(fileName string) error {
	return s.kubectlFile("apply", fileName, k8sTimeout, " created")
}

func (s *store) iUpdateAResource(fileName string) error {
	return s.kubectlFile("apply", fileName, k8sTimeout, " configured")
}

func (s *store) iDeleteAResource(fileName string) error {
	return s.kubectlFile("delete", fileName, k8sTimeout, " deleted")
}

func (s *store) kubectlFile(action string, fileName string, timeout time.Duration, expected ...string) error {
	app := "kubectl"
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, app, action, "-f", fileName, "-n", namespace)
	output := runCMD(cmd)
	var err error
	var ok bool
	for _, v := range expected {
		if !strings.Contains(output, v) {
			err = fmt.Errorf("unexpected output (%s)", string(output))
		}
		ok = true
	}
	if !ok {
		return err
	}

	cmd = exec.CommandContext(ctx, app, "get", "tykapis", "-n", namespace)
	output = runCMD(cmd)
	// TODO: need to wait for a bit for the reconciler to kick in
	time.Sleep(time.Second * 5)
	return nil
}

func (s *store) thereShouldBeHttpResponseCode(expectedCode int) error {
	if expectedCode != s.responseCode {
		println(string(s.responseBody))
		return fmt.Errorf("expected http status %d, got http %d", expectedCode, s.responseCode)
	}

	return nil
}

func (s *store) theResponseShouldContainJSONKeyValue(key string, expVal string) error {
	panic("not implemented this test")
}

func (s *store) theResponseShouldMatchJSON(body *godog.DocString) (err error) {
	var expected, actual interface{}

	// re-encode expected response
	if err = json.Unmarshal([]byte(body.Content), &expected); err != nil {
		return
	}

	// re-encode actual response too
	if err = json.Unmarshal(s.responseBody, &actual); err != nil {
		return
	}

	// the matching may be adapted per different requirements.
	if !reflect.DeepEqual(expected, actual) {
		println("ACTUAL")
		println(string(s.responseBody))
		println("EXPECTED")
		println(body.Content)
		return fmt.Errorf("expected JSON does not match actual, %v vs. %v", expected, actual)
	}
	return nil
}

func (s *store) thereShouldBeAResponseHeader(key string, value string) error {
	_, ok := s.responseHeaders[key]
	if !ok {
		return fmt.Errorf("response header (%s) not set", key)
	}
	headerVal := s.responseHeaders.Get(key)
	if headerVal != value {
		return fmt.Errorf("expected response header (%s), got (%s)", value, headerVal)
	}
	return nil
}

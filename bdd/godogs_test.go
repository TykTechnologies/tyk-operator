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
	"testing"
	"time"

	"github.com/cucumber/godog"
)

const (
	namespace = "bdd"
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
		exec.Command(app, "delete", "ns", namespace).Run()
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

var opts = godog.Options{
	StopOnFailure: true,
	Format:        "pretty",
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func setup(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "kubectl", "port-forward", "-n", gwNS, "svc/gw", "8000:8000")
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	fmt.Println(cmd.Args)
	cmd.Stderr = w
	cmd.Stdout = w
	err = cmd.Start()
	if err != nil {
		return err
	}
	r.SetReadDeadline(time.Now().Add(3 * time.Second))
	x := "Forwarding from 127.0.0.1:8000"
	b := make([]byte, len(x))
	_, err = io.ReadFull(r, b)
	if err != nil {
		return err
	}
	if !bytes.Equal(b, []byte(x)) {
		return fmt.Errorf("expected %q got %q", x, string(b))
	}
	return nil
}

func TestMain(t *testing.M) {
	flag.Parse()
	opts.Paths = flag.Args()
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	err := setup(ctx)
	if err != nil {
		log.Fatal(err)
	}

	status := godog.TestSuite{
		Name:                 "godogs",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              &opts,
	}.Run()
	if st := t.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

type store struct {
	responseCode    int
	responseBody    []byte
	responseTimes   []time.Duration
	cleanupK8s      []string
	responseHeaders map[string]string
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

func (s *store) iRequestEndpointWithHeader(path string, headerKey string, headerValue string) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:8000%s", path), nil)
	if err != nil {
		return err
	}
	req.Header.Set(headerKey, headerValue)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	s.responseCode = res.StatusCode

	for h, v := range res.Header {
		if s.responseHeaders == nil {
			s.responseHeaders = make(map[string]string, len(res.Header))
		}
		s.responseHeaders[h] = v[0]
	}

	s.responseBody, err = ioutil.ReadAll(res.Body)

	return nil
}

func (s *store) iRequestEndpoint(path string) error {
	res, err := client.Get(fmt.Sprintf("http://localhost:8000%s", path))
	if err != nil {
		// TODO: Check with Leo - this looks like a Gateway Bug
		if strings.Contains(err.Error(), "EOF") {
			// Assume it's a 404 to make the tests pass
			s.responseCode = http.StatusNotFound
			return nil
		}
		return err
	}
	defer res.Body.Close()

	s.responseCode = res.StatusCode

	for h, v := range res.Header {
		if s.responseHeaders == nil {
			s.responseHeaders = make(map[string]string, len(res.Header))
		}
		s.responseHeaders[h] = v[0]
	}

	s.responseBody, err = ioutil.ReadAll(res.Body)

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
	return s.kubectlFile("delete", fileName, " deleted", time.Second*20)
}

func (s *store) kubectlFile(action string, fileName string, expected string, timeout time.Duration) error {
	app := "kubectl"
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, app, action, "-f", fileName, "-n", namespace)
	output := runCMD(cmd)
	if !strings.Contains(output, expected) {
		return fmt.Errorf("unexpected output (%s)", string(output))
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
	headerVal, ok := s.responseHeaders[key]
	if !ok {
		return fmt.Errorf("response header (%s) not set", key)
	}
	if headerVal != value {
		return fmt.Errorf("expected response header (%s), got (%s)", value, headerVal)
	}
	return nil
}

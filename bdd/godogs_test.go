package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"github.com/TykTechnologies/tyk-operator/bdd/k8sutil"
	"github.com/cenkalti/backoff/v4"
	"github.com/cucumber/godog"
)

const (
	namespace      = "bdd"
	k8sTimeout     = time.Second * 10
	reconcileDelay = time.Second * 1
)

var gwNS = fmt.Sprintf("tyk%s-control-plane", os.Getenv("TYK_MODE"))

func runCMD(cmd *exec.Cmd) string {
	a := fmt.Sprint(cmd.Args)

	fmt.Println(a)

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("failed %s with %v : %s", a, err, string(output)))
	}

	return string(output)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		k8sutil.DeleteNS(context.Background(), namespace)
		err := k8sutil.CreateNS(context.Background(), namespace)
		if err != nil {
			panic(err)
		}

		err = k8sutil.Create(context.Background(), "./custom_resources/workaround.yaml", namespace)
		if err != nil {
			panic(err)
		}
	})
}

var opts = &godog.Options{
	StopOnFailure: true,
	Format:        "pretty",
	Tags:          "~@undone",
}
var gatewayURL = "http://localhost:8080"

func init() {
	godog.BindCommandLineFlags("godog.", opts)
}

func TestMain(t *testing.M) {
	flag.Parse()

	opts.Paths = flag.Args()

	kill, err := k8sutil.Init(gwNS, os.Getenv("TYK_MODE"))
	if err != nil {
		log.Fatal(err)
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

	err = kill()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(status)
}

type store struct {
	responseCode    int
	responseBody    []byte
	responseTimes   []time.Duration
	created         map[string]struct{}
	responseHeaders http.Header
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	s := &store{
		created: map[string]struct{}{},
	}

	ctx.AfterScenario(func(sc *godog.Scenario, err error) {
		for fileName := range s.created {
			ctx, cancel := context.WithTimeout(context.Background(), k8sTimeout)
			defer cancel()
			wait(reconcileDelay)(
				k8sutil.Delete(ctx, fileName, namespace),
			)
		}
	})

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

func (s *store) iRequestEndpointWithHeaderTimes(path, headerKey, headerValue string, times int) error {
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
	validate func(*http.Response) error,
) error {
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

		res, err := k8sutil.Do(req)
		if err != nil {
			fmt.Println("==========> Error making client call ", err)
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

func (s *store) iRequestEndpointWithHeader(path, headerKey, headerValue string) error {
	return call(
		http.MethodGet,
		createURL(path),
		func() io.Reader { return nil },
		func(h *http.Request) {
			h.Header.Set(headerKey, headerValue)
		},
		func(res *http.Response) error {
			var err error

			s.responseCode = res.StatusCode
			s.responseHeaders = res.Header.Clone()
			s.responseBody, err = io.ReadAll(res.Body)

			if err != nil {
				return err
			}

			return nil
		},
	)
}

func createURL(path string) string {
	return gatewayURL + path
}

func (s *store) iRequestEndpoint(path string) error {
	return call(
		http.MethodGet,
		createURL(path),
		func() io.Reader { return nil },
		func(h *http.Request) {},
		func(h *http.Response) error {
			var err error

			s.responseCode = h.StatusCode
			s.responseHeaders = h.Header.Clone()
			s.responseBody, err = io.ReadAll(h.Body)

			if err != nil {
				return err
			}

			return nil
		},
	)
}

func (s *store) thereIsAResource(fileName string) error {
	s.created[fileName] = struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), k8sTimeout)

	defer cancel()

	return wait(reconcileDelay)(
		k8sutil.Create(ctx, fileName, namespace),
	)
}

func (s *store) iCreateAResource(fileName string) error {
	s.created[fileName] = struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), k8sTimeout)

	defer cancel()

	return wait(reconcileDelay)(
		k8sutil.Create(ctx, fileName, namespace),
	)
}

func (s *store) iUpdateAResource(fileName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), k8sTimeout)

	defer cancel()

	return wait(reconcileDelay)(
		k8sutil.Configure(ctx, fileName, namespace),
	)
}

func (s *store) iDeleteAResource(fileName string) error {
	delete(s.created, fileName)

	ctx, cancel := context.WithTimeout(context.Background(), k8sTimeout)

	defer cancel()

	return wait(reconcileDelay)(
		k8sutil.Delete(ctx, fileName, namespace),
	)
}

func wait(ts time.Duration) func(err error) error {
	return func(err error) error {
		if err != nil {
			return err
		}

		time.Sleep(ts)

		cmd := exec.CommandContext(context.Background(), "kubectl", "get", "tykapis", "-n", namespace)

		fmt.Println(runCMD(cmd))

		return nil
	}
}

func (s *store) thereShouldBeHttpResponseCode(expectedCode int) error {
	if expectedCode != s.responseCode {
		println(string(s.responseBody))
		return fmt.Errorf("expected http status %d, got http %d", expectedCode, s.responseCode)
	}

	return nil
}

func (s *store) theResponseShouldContainJSONKeyValue(key, expVal string) error {
	m := map[string]interface{}{}

	if err := json.Unmarshal(s.responseBody, &m); err != nil {
		return err
	}

	got := fmt.Sprint(m[key])
	if got != expVal {
		return fmt.Errorf("expected %q got %q", expVal, got)
	}

	return nil
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

func (s *store) thereShouldBeAResponseHeader(key, value string) error {
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

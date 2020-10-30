package main

import (
	"errors"

	"github.com/cucumber/godog"
)

type store struct {
	responseCode int
	cleanupK8s   []string
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	s := store{}

	ctx.Step(`^there is a (\S+) resource`, s.thereIsAResource)
	ctx.Step(`^i update a (\S+) resource`, s.iUpdateAResource)
	ctx.Step(`^i request (\S+) endpoint$`, s.iRequestEndpoint)
	ctx.Step(`^there should be a (\d+) http response code$`, s.thereShouldBeHttpResponseCode)
}

func (s *store) iRequestEndpoint(path string) error {
	//myCommand := []string{
	//	"run",
	//	"mycurl",
	//	"-it",
	//	"--rm",
	//	"--image",
	//	"radial/busyboxplus:curl",
	//	fmt.Sprintf("http://gw.tykpro-control-plant.svc:8000%s", path),
	//}
	//
	//cmd := exec.Command("kubectl", myCommand...)
	//output, err := cmd.Output()
	//if err != nil {
	//	return err
	//}
	//println(string(output))

	//res, err := http.Get(path)
	//if err != nil {
	//	return err
	//}

	s.responseCode = 200

	return nil
}

func (s *store) thereIsAResource(arg1 string) error {
	//app := "kubectl"
	//
	//cmd := exec.Command(app, "apply", "-f", arg1)
	//output, err := cmd.Output()
	//
	//println(string(output))
	//// TODO: do something to test that the resource was actually created
	//
	//if err != nil {
	//	return err
	//}

	return nil
}

func (s *store) iUpdateAResource(arg1 string) error {
	return s.thereIsAResource(arg1)
}

func (s *store) thereShouldBeHttpResponseCode(expectedCode int) error {
	if expectedCode != 200 {
		return errors.New("unexpected response code")
	}

	return nil
}

package main

import (
	"os/exec"

	"github.com/cucumber/godog"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^there is a (\s+) resource$`, thereIsAResource)
}

func thereIsAResource(arg1 string) error {
	app := "kubectl"

	cmd := exec.Command(app, "apply", "-f", arg1)
	_, err := cmd.Output()

	// TODO: do something to test that the resource was actually created

	if err != nil {
		return err
	}

	return nil
}

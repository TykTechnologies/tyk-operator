package integration

import (
	"log"
	"os"
	"testing"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"sigs.k8s.io/e2e-framework/pkg/env"
)

var (
	e       environmet.Env
	testenv env.Environment
)

func TestMain(t *testing.M) {
	e.Parse()
	if e.Mode == "" {
		log.Fatal("Missing TYK_MODE")
	}
	testenv = env.New()
	testenv.Setup(
		setupk8s,
		setupTyk,
		setupE2E,
		setupMultiTenancy,
	).Finish(
		teardownMultiTenancy,
		teardownE2E,
		teardownTyk,
		teardownk8s,
	)
	os.Exit(testenv.Run(t))
}

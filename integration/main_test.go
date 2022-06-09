package integration

import (
	"log"
	"os"
	"testing"

	"github.com/TykTechnologies/tyk-operator/integration/setup"

	"github.com/TykTechnologies/tyk-operator/integration/common"
	"sigs.k8s.io/e2e-framework/pkg/env"
)

var testenv env.Environment

func TestMain(t *testing.M) {
	common.Env.Parse()

	if common.Env.Mode == "" {
		log.Fatal("Missing TYK_MODE")
	}

	testenv = env.New()

	testenv.Setup(
		setup.Kubernetes,
		setup.Tyk,
		setup.SetupE2E,
		setup.MultiTenancy,
	).Finish(
		setup.TeardownMultiTenancy,
		setup.TeardownE2E,
		setup.TeardownTyk,
		setup.TeardownKubernetes,
	).BeforeEachTest(
		setup.CreateNamespace,
	).AfterEachTest(
		setup.DeleteNamespace,
	)

	os.Exit(testenv.Run(t))
}

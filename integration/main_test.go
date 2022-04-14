package integration

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var (
	e       environmet.Env
	testenv env.Environment
)

type ctxKey string

const (
	ctxNSKey     ctxKey = "namespaceName"
	ctxApiName   ctxKey = "apiName"
	ctxOpCtxName ctxKey = "opCtxName"
)

const (
	defaultWaitTimeout  = 1 * time.Minute
	defaultWaitInterval = 1 * time.Second
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
	).BeforeEachTest(
		createNamespace,
	).AfterEachTest(
		deleteNamespace,
	)

	os.Exit(testenv.Run(t))
}

func createNamespace(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
	name := envconf.RandomName("tyk-operator", 32)

	ctx = context.WithValue(ctx, ctxNSKey, name)

	nsObj := v1.Namespace{}
	nsObj.Name = name
	return ctx, cfg.Client().Resources().Create(ctx, &nsObj)
}

func deleteNamespace(ctx context.Context, envconf *envconf.Config, t *testing.T) (context.Context, error) {
	name := ctx.Value(ctxNSKey)

	nsObj := v1.Namespace{}
	nsObj.Name = name.(string)
	return ctx, envconf.Client().Resources().Delete(ctx, &nsObj)
}

package migrate

import (
	"os"

	"github.com/TykTechnologies/tyk-operator/pkg/dashboard_client"
	"github.com/TykTechnologies/tyk-operator/pkg/environmet"
	"github.com/go-logr/zapr"
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

var CMD = cli.Command{
	Name:  "migrate",
	Usage: "migrates api's and policies from tyk dashboard to k8s resources",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "url,d",
			Usage:  "URL to the tyk dashboard",
			EnvVar: environmet.TykURL,
		},
		cli.StringFlag{
			Name:   "auth,a",
			Usage:  "Auth token",
			EnvVar: environmet.TykAuth,
		},
	},
	Action: func(ctx *cli.Context) error {
		c := zap.NewProductionConfig()
		c.DisableStacktrace = true
		xl, err := c.Build(
			zap.WithCaller(false),
		)
		if err != nil {
			return err
		}
		e := environmet.Env{
			Auth: ctx.String("auth"),
			URL:  ctx.String("url"),
		}
		lg := zapr.NewLogger(xl)
		client := dashboard_client.NewClient(lg, e)
		xl.Info("Fetching api's")
		apis, err := client.Api().All()
		if err != nil {
			xl.Error("Failed to get apis", zap.Error(err))
			return err
		}
		xl.Info("received apis ", zap.Int("count", len(apis)))
		xl.Info("Fetching policies")
		policies, err := client.SecurityPolicy().All()
		if err != nil {
			xl.Error("Failed to get polices", zap.Error(err))
			return err
		}
		xl.Info("received polices ", zap.Int("count", len(policies)))
		return Build(os.Stdout, e.Namespace, apis, policies)
	},
}

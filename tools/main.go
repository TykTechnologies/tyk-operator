package main

import (
	"log"
	"os"

	"github.com/TykTechnologies/tyk-operator/tools/migrate"
	"github.com/urfave/cli"
)

func main() {
	a := cli.NewApp()
	a.Name = "Tyk Operator tools"
	a.Usage = "Tools for smooth sailing with tyk operator"
	a.Commands = cli.Commands{
		migrate.CMD,
	}
	if err := a.Run(os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

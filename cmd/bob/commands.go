package main

import "github.com/urfave/cli/v2"

var configFlag = cli.StringFlag{
	Name:    "config",
	Aliases: []string{"c"},
	Usage:   "Path to a bob.json config file",
}

var buildKubernetesCmd = cli.Command{
	Name:  "build-kubernetes",
	Usage: "build-kubernetes FLAGS",
	Flags: []cli.Flag{
		&configFlag,
	},
}

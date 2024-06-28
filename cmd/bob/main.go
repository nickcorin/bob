package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var (
	buildDir = flag.String(
		"buildDir",
		"",
		"Path to the root of the build directory.",
	)

	config = flag.String(
		"config",
		"",
		"Path to a bob.json configuration file.",
	)
)

// TODO: Edit Makefile templates to reflect new bob changes. The generate target
// is going to be broken.

func main() {
	app := cli.NewApp()

	app.Commands = []*cli.Command{
		&buildKubernetesCmd,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Printf("Usage: %s <options>\n", os.Args[0])
	fmt.Printf("\n")
	fmt.Printf("Bob the Builder, can he fix it? NO! But he can generate some code.\n")
	fmt.Printf("\n")
	fmt.Printf("Either --buildDir or --config must be set.\n")
	fmt.Printf("\n")
	fmt.Printf("Options:\n")
	flag.PrintDefaults()
	fmt.Printf("\n")
}

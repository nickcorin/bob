package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nickcorin/adventech/bob"
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
	flag.Parse()

	if *config != "" {
		if err := bob.GenerateService(*config); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	if *buildDir != "" {
		if err := bob.GenerateDockerCompose(*buildDir); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}

	printUsage()
	os.Exit(2)
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

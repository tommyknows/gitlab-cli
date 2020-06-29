package main

import (
	"os"

	"github.com/tommyknows/gitlab-cli/cmd"
	"github.com/tommyknows/gitlab-cli/pkg/log"
)

func main() {
	// for now, debug logging is hardcoded because we're in
	// very early stages. TODO: change that, add a flag.
	log.Setup("debug")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

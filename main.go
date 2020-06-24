package main

import (
	"os"

	"github.com/tommyknows/gitlab-cli/cmd"
	"github.com/tommyknows/gitlab-cli/pkg/log"
)

func main() {
	log.Setup("debug")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tommyknows/gitlab-cli/api/config"
)

func Execute() error {
	return gitlabCLI().Execute()
}

var (
	// version is set at compile time to either the commit SHA or tag
	// of the current build. it is used by Cobra to generate the "--version"
	// flag, if set
	version string
)

func gitlabCLI() *cobra.Command {
	var cfgFile string
	cfg := new(config.Config)

	var useConfigContext bool

	configDefaultPath := ""
	if home := homeDir(); home != "" {
		configDefaultPath = filepath.Join(home, ".gitlab-cli.json")
	}

	cmd := &cobra.Command{
		Version: version,
		Use:     "gitlab-cli",
		Short:   "gitlab-cli allows interacting with Gitlab",
		Long:    `TODO`, // TODO
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			c, err := config.Load(cfgFile, useConfigContext)
			if err != nil {
				return err
			}

			*cfg = *c
			return err
		},
		PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
			return cfg.Write()
		},
	}

	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", configDefaultPath, "config file location")
	// TODO: find a better name for this
	cmd.PersistentFlags().BoolVarP(&useConfigContext, "use-config-context", "u", false, "use the context of the config instead of a possible local one")
	cmd.AddCommand(
		newContextCommand(cfg),
		newInstanceCommand(cfg),
		newProjectCommand(cfg),
	)

	return cmd
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

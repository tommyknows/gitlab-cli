package cmd

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tommyknows/gitlab-cli/api/config"
	"github.com/tommyknows/gitlab-cli/pkg/log"
)

func Execute() error {
	return gitlabCLI().Execute()
}

type command struct {
	name string
	abbr []string
}

var (
	listSub = command{
		name: "list",
		abbr: []string{"ls"},
	}
	createSub = command{
		name: "create",
		abbr: []string{"cr", "new", "add"}, // TODO: is this a good idea?
	}
	deleteSub = command{
		name: "delete",
		abbr: []string{"del", "rm"},
	}
)

func (c *command) Usage(s string) string {
	return c.name + " " + s
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
		configDefaultPath = filepath.Join(home, ".gitlab-cli.yml")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan
		log.Debugf("canceling context ,stopping all operations")
		cancel()
	}()

	cmd := &cobra.Command{
		Version: version,
		Use:     "gitlab-cli",
		Short:   "gitlab-cli allows interacting with Gitlab",
		Long:    `TODO`, // TODO
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			c, err := config.Load(cfgFile, useConfigContext)
			if err != nil {
				return errors.Wrapf(err, "could not load config file")
			}

			*cfg = *c

			return nil
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
		newProjectCommand(ctx, cfg),
	)

	return cmd
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

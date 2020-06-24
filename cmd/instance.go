package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tommyknows/gitlab-cli/api/config"
)

func newInstanceCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Short:   "manage instance connections of the config file",
		Aliases: []string{"inst"},
		Use:     "instance",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "add [SERVER] [TOKEN]",
			Short: "login to a Gitlab instance. if none is given, uses gitlab.com",
			Args:  cobra.RangeArgs(1, 2),
			RunE: func(_ *cobra.Command, args []string) error {
				var token, server string
				if len(args) == 1 {
					server = "gitlab.com"
					token = args[0]
				} else {
					server = args[0]
					token = args[1]
				}

				cfg.Instances[server] = &config.InstanceConfig{
					Authentication: &config.Authentication{
						Type: config.Token,
						TokenAuthentication: &config.TokenAuthentication{
							Token: token,
						},
					},
				}

				if _, ok := cfg.Contexts[server]; ok {
					fmt.Printf("context with name %v already exists, not modifying\n", server)
				} else {
					cfg.Contexts[server] = &config.Context{
						InstanceName: server,
						Group:    "",
					}
				}

				cfg.CurrentContext = server

				fmt.Printf("added login as context %v and set as current context", server)

				return nil
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "list all instances in the config file",
			Args:  cobra.NoArgs,
			RunE: func(_ *cobra.Command, args []string) error {
				return printInstances(cfg)
			},
		},
		&cobra.Command{
			Use:   "clean",
			Short: "clean the config file, pruning instances that are not referenced in a context",
			Args:  cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
			NextInstance:
				for instance := range cfg.Instances {
					for _, ctx := range cfg.Contexts {
						if ctx.InstanceName == instance {
							continue NextInstance
						}
					}

					delete(cfg.Instances, instance)
				}

				return nil
			},
		},
	)
	return cmd
}

func printInstances(cfg *config.Config) error {
	w := tabwriter.NewWriter(os.Stdout, 1, 8, 2, ' ', 0)

	fmt.Fprint(w, "\tname\tauthentication type\n")
	fmt.Fprint(w, "\t----\t-------------------\n")

	ctx, err := cfg.GetCurrentContext()
	if err != nil {
		return err
	}

	for name, instance := range cfg.Instances {
		var current string
		if name == ctx.InstanceName {
			current = "*"
		}
		fmt.Fprintf(w, "%s\t%v\t%v\n", current, name, instance.Authentication.Type)
	}

	return w.Flush()

}

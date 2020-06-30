package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tommyknows/gitlab-cli/api/config"
)

func newContextCommand(cfg *config.Config) *cobra.Command {
	c := &cobra.Command{
		Long: `the context subcommands allows to list, add, change or delete contexts that are 
specified in the config file.

A context specifies the mapping between a namespace and an instance, allowing to 
easily "switch" projects upstream. 

If the context is a remote command or the "--use-config-context" flag is set, the 
current context will be used to specify the "root" for all operations. 

Some commands may allow to pass a project as an argument, which may have a 
different behaviour depending on how the project is specified. The project 
definitions work similarly to Unix / Linux directories: 
	If the project starts with a "/", the path will be deemed to be absolute, while
if it does not start with a "/", it will be considered to be relative to the 
current namespace.`,
		Short:   "work with the contexts",
		Aliases: []string{"contexts", "ctx", "ctxs"},
		Use:     "context",
	}

	c.AddCommand(
		&cobra.Command{
			Use:     "list",
			Short:   "list all available contexts",
			Aliases: []string{"ls", "l"},
			Args:    cobra.NoArgs,
			RunE: func(_ *cobra.Command, args []string) error {
				return printContexts(cfg)
			},
		},
		&cobra.Command{
			Use:     "create [name] [instance] [namespace/project]",
			Short:   "create a context that is tied to an instance, with an optional group",
			Args:    cobra.RangeArgs(2, 3),
			Aliases: []string{"cr"},
			RunE: func(_ *cobra.Command, args []string) error {
				var name, instance, ns string
				name = args[0]
				instance = args[1]

				if len(args) == 3 {
					ns = args[2]
				}

				cfg.Contexts[name] = &config.Context{
					Namespace:    ns,
					InstanceName: instance,
				}

				if _, ok := cfg.Instances[instance]; !ok {
					fmt.Println("Instance", instance, "is not specified in config")
				}
				return nil
			},
		},
		&cobra.Command{
			Use:          "switch [name]",
			Short:        "switch to a context",
			Aliases:      []string{"sw"},
			SilenceUsage: true,
			Args:         cobra.ExactArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				if _, ok := cfg.Contexts[args[0]]; !ok {
					return errors.Errorf("no such context: %q", args[0])
				}
				cfg.CurrentContext = args[0]

				return nil
			},
		},
		&cobra.Command{
			Use:     "delete [name]",
			Short:   "delete a context",
			Aliases: []string{"del", "rm"},
			Args:    cobra.ExactArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				delete(cfg.Contexts, args[0])
				return nil
			},
		},
		&cobra.Command{
			Use:   "clean",
			Short: "clean the config file, pruning contexts that reference an instance that does not exist",
			Args:  cobra.NoArgs,
			RunE: func(_ *cobra.Command, _ []string) error {
				for name, ctx := range cfg.Contexts {
					if _, ok := cfg.Instances[ctx.InstanceName]; !ok {
						delete(cfg.Contexts, name)
					}
				}

				return nil
			},
		},
	)
	return c
}

func printContexts(c *config.Config) error {
	w := tabwriter.NewWriter(os.Stdout, 1, 8, 2, ' ', 0)

	fmt.Fprint(w, "\tname\tinstance\tnamespace\n")
	fmt.Fprint(w, "\t----\t--------\t---------\n")

	for name, ctx := range c.Contexts {
		var current string
		if name == c.CurrentContext {
			current = "*"
		}

		fmt.Fprintf(w, "%s\t%v\t%v\t%v\n", current, name, ctx.InstanceName, ctx.Namespace)
	}

	return w.Flush()
}

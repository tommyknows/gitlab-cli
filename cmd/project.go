package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tommyknows/gitlab-cli/api/config"
	"github.com/tommyknows/gitlab-cli/pkg/gitlab"
	"github.com/tommyknows/gitlab-cli/pkg/log"
)

func newProjectCommand(cfg *config.Config) *cobra.Command {
	c := &cobra.Command{
		Short:   "work with projects",
		Aliases: []string{"pr", "proj", "projects"},
		Use:     "project",
	}

	useCtx := &cobra.Command{
		Use:          "use-context [project] [context-name]",
		Short:        "use specified project / group as context ",
		SilenceUsage: true,
		Long: "use the specified project / group as context. If context name is not given, " +
			"current context will be overwritten. If project is an absolute path," +
			"it will be added to the currently active project.",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(_ *cobra.Command, args []string) error {
			// the project command doesn't really make sense if a concrete git repo.
			cfg.PreferConfigContext = true

			proj := args[0]

			currentCtx, err := cfg.GetCurrentContext()
			if err != nil {
				return errors.Wrapf(err, "could not get active context")
			}

			group := currentCtx.Group
			if currentCtx.Group == "" {
				group = currentCtx.User
			}

			group = getAbsoluteGroupPath(group, proj)

			if len(args) == 1 {
				currentCtx.Group = group
				return nil
			}

			ctxName := args[1]
			cfg.Contexts[ctxName] = &config.Context{
				Group:        group,
				InstanceName: currentCtx.InstanceName,
			}

			return nil
		},
	}

	c.AddCommand(
		newProjectListCommand(cfg),
		newProjectCloneCommand(cfg),
		useCtx)
	return c
}

func newProjectCloneCommand(cfg *config.Config) *cobra.Command {
	var (
		recursive bool
		depth     int

		clone = &cobra.Command{
			Use:          "clone [proj]",
			SilenceUsage: true,
			// TODO
			Short:   "clone a group or project recursively, creating the necessary folders",
			Aliases: []string{"cl"},
			Args:    cobra.RangeArgs(0, 1),
			RunE: func(_ *cobra.Command, args []string) error {
				// the project command doesn't really make sense if a concrete git repo.
				cfg.PreferConfigContext = true

				cctx, err := cfg.GetCurrentContext()
				if err != nil {
					return err
				}

				var group string
				if len(args) == 1 {
					group = args[0]
				}

				group = getAbsoluteGroupPath(cctx.Group, group)

				client, err := cctx.WithGroup(group).GitlabClient()
				if err != nil {
					return errors.Wrapf(err, "could not get gitlab client")
				}

				log.Infof("fetching projects...")

				rootProj, err := client.GetProjects(false)
				if err != nil {
					return errors.Wrapf(err, "could not get group")
				}

				return gitlab.Walk(rootProj, gitlab.Clone(rootProj.FullPath()))
			},
		}
	)

	clone.Flags().BoolVarP(&recursive, "recursive", "r", false, "list recursively")
	clone.Flags().IntVarP(&depth, "depth", "d", -1, "depth to list recursively. -1 means infinite")
	return clone
}

func newProjectListCommand(cfg *config.Config) *cobra.Command {
	var (
		recursive       bool
		depth           int
		showDescription bool
		showAll         bool

		list = &cobra.Command{
			Use:          "list [proj]",
			SilenceUsage: true,
			// TODO
			Short:   "list projects in the current group",
			Aliases: []string{"ls"},
			Args:    cobra.RangeArgs(0, 1),
			RunE: func(_ *cobra.Command, args []string) error {
				// TODO: move this to PersistentPreRun in Project command. Couldn't get it to work.
				// the project command doesn't really make sense if a concrete git repo.
				cfg.PreferConfigContext = true

				cctx, err := cfg.GetCurrentContext()
				if err != nil {
					return err
				}

				var group string
				if len(args) == 1 {
					group = args[0]
				}

				group = getAbsoluteGroupPath(cctx.Group, group)

				client, err := cctx.WithGroup(group).GitlabClient()
				if err != nil {
					return errors.Wrapf(err, "could not get gitlab client")
				}

				log.Infof("fetching projects...")

				rootProj, err := client.GetProjects(showAll)
				if err != nil {
					return errors.Wrapf(err, "could not get group")
				}

				fmt.Printf("%v", gitlab.PrintProject(rootProj, gitlab.PrintOptions{
					PrintArchived:    showAll,
					PrintDescription: showDescription,
					Depth:            depth,
				}))
				return nil
			},
		}
	)

	list.Flags().BoolVarP(&recursive, "recursive", "r", false, "list recursively")
	list.Flags().IntVarP(&depth, "depth", "d", 0, "depth to list recursively. 0 means infinite")
	list.Flags().BoolVar(&showDescription, "desc", false, "show description of projects too")
	list.Flags().BoolVarP(&showAll, "all", "a", false, "show all projects, including archived ones")

	return list
}

// getAbsoluteGroupPath checks if the given newGroup is a relative or
// absolute path. If it is an absolute path, this is returned. If it is
// relative, it is appended to the currentGroup.
func getAbsoluteGroupPath(currentGroup, newGroup string) string {
	if strings.HasPrefix(newGroup, "/") {
		return strings.TrimSuffix(strings.TrimPrefix(newGroup, "/"), "/")
	}
	return strings.TrimSuffix(strings.TrimPrefix(path.Join(currentGroup, newGroup), "/"), "/")
}

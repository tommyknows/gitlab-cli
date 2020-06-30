package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tommyknows/gitlab-cli/api/config"
	"github.com/tommyknows/gitlab-cli/pkg/gitlab"
	"github.com/tommyknows/gitlab-cli/pkg/log"
)

func newProjectCommand(ctx context.Context, cfg *config.Config) *cobra.Command {
	c := &cobra.Command{
		Short: "work with projects",
		Long: `The project subcommand allows to work with projects. It DOES NOT make a 
distinction between groups, user (both are considered to be "namespaces" in gitlab terms),
or actual projects (git repository).`,
		Aliases: []string{"pr", "proj", "projects"},
		Use:     "project",
	}

	useCtx := &cobra.Command{
		Use:          "use-context [project] [context-name]",
		Aliases:      []string{"use-ctx"},
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

			namespace := getAbsoluteGroupPath(currentCtx.Namespace, proj)

			if len(args) == 1 {
				currentCtx.Namespace = namespace
				return nil
			}

			ctxName := args[1]
			cfg.Contexts[ctxName] = &config.Context{
				Namespace:    namespace,
				InstanceName: currentCtx.InstanceName,
			}

			return nil
		},
	}

	c.AddCommand(
		newProjectListCommand(ctx, cfg),
		newProjectCloneCommand(ctx, cfg),
		useCtx)
	return c
}

func newProjectCloneCommand(ctx context.Context, cfg *config.Config) *cobra.Command {
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
					return errors.Wrapf(err, "could not get current context")
				}

				var namespace string
				if len(args) == 1 {
					namespace = args[0]
				}

				var skipRoot bool
				if strings.HasSuffix(namespace, "/") {
					skipRoot = true
				}

				namespace = getAbsoluteGroupPath(cctx.Namespace, namespace)

				client, err := cctx.WitNamespace(namespace).GitlabClient()
				if err != nil {
					return errors.Wrapf(err, "could not get gitlab client")
				}

				log.Infof("fetching projects...")

				rootProj, err := client.GetProjects(ctx, false)
				if err != nil {
					return errors.Wrapf(err, "could not get namespace or project %s", namespace)
				}

				rootPath := rootProj.Namespace()
				if skipRoot {
					rootPath = rootProj.FullPath()
				}

				clone, err := gitlab.Clone(rootPath, skipRoot, cctx.Authentication())
				if err != nil {
					return errors.Wrapf(err, "could not setup clone environment")
				}
				// TODO: implement canceling the context by stopping the binary
				return gitlab.WalkConcurrent(ctx, rootProj, clone)
			},
		}
	)

	clone.Flags().BoolVarP(&recursive, "recursive", "r", false, "list recursively")
	clone.Flags().IntVarP(&depth, "depth", "d", -1, "depth to list recursively. -1 means infinite")
	return clone
}

func newProjectListCommand(ctx context.Context, cfg *config.Config) *cobra.Command {
	var (
		depth           int
		showDescription bool
		showAll         bool

		list = &cobra.Command{
			Use:          listSub.Usage("[proj]"),
			SilenceUsage: true,
			Short:        "list projects in the current group", // TODO
			Aliases:      listSub.abbr,
			Args:         cobra.RangeArgs(0, 1),
			RunE: func(_ *cobra.Command, args []string) error {
				// TODO: move this to PersistentPreRun in Project command. Couldn't get it to work.
				// the project command doesn't really make sense if a concrete git repo.
				cfg.PreferConfigContext = true

				cctx, err := cfg.GetCurrentContext()
				if err != nil {
					return errors.Wrapf(err, "could not get current context")
				}

				var namespace string
				if len(args) == 1 {
					namespace = args[0]
				}

				namespace = getAbsoluteGroupPath(cctx.Namespace, namespace)

				client, err := cctx.WitNamespace(namespace).GitlabClient()
				if err != nil {
					return errors.Wrapf(err, "could not get gitlab client")
				}

				log.Infof("fetching projects...")

				rootProj, err := client.GetProjects(ctx, showAll)
				if err != nil {
					return errors.Wrapf(err, "could not get namespace or project %s", namespace)
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

	list.Flags().IntVarP(&depth, "depth", "d", 1, "depth to list recursively. 0 means infinite")
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

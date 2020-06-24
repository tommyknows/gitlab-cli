package cmd

import "github.com/spf13/cobra"

func clone() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "clone a repository or a whole group",
		Long:  `TODO`, // TODO
	}
	// TODO: short hand flag would be nice...
	cloneRecursive := cmd.Flags().Bool("recursive", false, "clone group recursively")

	// TODO
	cmd.RunE = func(_ *cobra.Command, args []string) error {
		if *cloneRecursive {
			return nil
		}
		return nil
	}

	return cmd
}

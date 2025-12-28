package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <workspace>",
	Short: "Open or attach to an existing workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		noResume, _ := cmd.Flags().GetBool("no-resume")

		mgr, err := newManager()
		if err != nil {
			return err
		}

		if err := mgr.OpenWorkspace(cmd.Context(), id, !noResume); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "opened workspace %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
	openCmd.Flags().Bool("no-resume", false, "Do not resume Claude Code session")
}

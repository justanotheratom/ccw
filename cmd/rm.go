package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <workspace>",
	Short: "Remove a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("rm: not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)

	rmCmd.Flags().BoolP("force", "f", false, "Force removal even if branch not merged")
	rmCmd.Flags().Bool("keep-branch", false, "Keep the git branch")
	rmCmd.Flags().Bool("keep-worktree", false, "Keep the worktree (just unregister)")
	rmCmd.Flags().Bool("yes", false, "Skip confirmation prompts")
}

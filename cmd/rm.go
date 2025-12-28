package cmd

import (
	"fmt"

	"github.com/ccw/ccw/internal/workspace"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <workspace>",
	Short: "Remove a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")
		keepBranch, _ := cmd.Flags().GetBool("keep-branch")
		keepWorktree, _ := cmd.Flags().GetBool("keep-worktree")

		mgr, err := newManager()
		if err != nil {
			return err
		}

		err = mgr.RemoveWorkspace(cmd.Context(), id, workspace.RemoveOptions{
			Force:        force,
			KeepBranch:   keepBranch,
			KeepWorktree: keepWorktree,
		})
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "removed workspace %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)

	rmCmd.Flags().BoolP("force", "f", false, "Force removal even if branch not merged")
	rmCmd.Flags().Bool("keep-branch", false, "Keep the git branch")
	rmCmd.Flags().Bool("keep-worktree", false, "Keep the worktree (just unregister)")
	rmCmd.Flags().Bool("yes", false, "Skip confirmation prompts")
}

package cmd

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <workspace>",
	Short: "Show detailed information about a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		mgr, err := newManager()
		if err != nil {
			return err
		}

		status, err := mgr.WorkspaceInfo(cmd.Context(), id)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Workspace:\t%s\n", status.ID)
		fmt.Fprintf(w, "Repo:\t%s\n", status.Workspace.RepoPath)
		fmt.Fprintf(w, "Worktree:\t%s\n", status.Workspace.WorktreePath)
		fmt.Fprintf(w, "Branch:\t%s\n", status.Workspace.Branch)
		fmt.Fprintf(w, "Base:\t%s\n", status.Workspace.BaseBranch)
		fmt.Fprintf(w, "Claude Session:\t%s\n", status.Workspace.ClaudeSession)
		fmt.Fprintf(w, "Tmux Session:\t%s\n", status.Workspace.TmuxSession)
		fmt.Fprintf(w, "Session Alive:\t%t\n", status.SessionAlive)
		fmt.Fprintf(w, "Created:\t%s\n", status.Workspace.CreatedAt.Format(time.RFC3339))
		fmt.Fprintf(w, "Last Accessed:\t%s\n", status.Workspace.LastAccessedAt.Format(time.RFC3339))
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

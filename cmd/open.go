package cmd

import (
	"errors"
	"fmt"

	"github.com/ccw/ccw/internal/workspace"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <workspace>",
	Short: "Open or attach to an existing workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		noResume, _ := cmd.Flags().GetBool("no-resume")
		focusExisting, _ := cmd.Flags().GetBool("focus")
		forceAttach, _ := cmd.Flags().GetBool("attach")

		mgr, err := newManager()
		if err != nil {
			return err
		}

		if err := mgr.OpenWorkspace(cmd.Context(), id, workspace.OpenOptions{
			ResumeClaude:  !noResume,
			FocusExisting: focusExisting,
			ForceAttach:   forceAttach,
		}); err != nil {
			if errors.Is(err, workspace.ErrWorkspaceAlreadyOpen) {
				return fmt.Errorf("workspace %s is already open (use --focus to focus the existing window)", id)
			}
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "opened workspace %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
	openCmd.Flags().Bool("no-resume", false, "Do not resume Claude Code session")
	openCmd.Flags().Bool("focus", false, "Focus existing window if the workspace is already open")
	openCmd.Flags().Bool("attach", false, "Force attach even when not running in a TTY")
}

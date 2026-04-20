package cmd

import (
	"errors"
	"fmt"

	"github.com/ccw/ccw/internal/workspace"
	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close [workspace]",
	Short: "Close a workspace session (defaults to the current one)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := newManager()
		if err != nil {
			return err
		}

		var id string
		if len(args) == 1 {
			id = args[0]
		} else {
			curID, _, err := mgr.FindCurrent(cmd.Context())
			if err != nil {
				if errors.Is(err, workspace.ErrNoCurrentWorkspace) {
					return fmt.Errorf("not inside a ccw workspace; pass a workspace id (ccw close <workspace>) or cd into one")
				}
				return err
			}
			id = curID
		}

		fmt.Fprintf(cmd.OutOrStdout(), "closing workspace %s\n", id)

		if err := mgr.CloseWorkspace(cmd.Context(), id); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "closed workspace %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(closeCmd)
}

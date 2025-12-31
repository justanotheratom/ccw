package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ccw/ccw/internal/workspace"
	"github.com/spf13/cobra"
)

var staleCmd = &cobra.Command{
	Use:   "stale",
	Short: "List workspaces with merged branches",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		remove, _ := cmd.Flags().GetBool("rm")
		force, _ := cmd.Flags().GetBool("force")
		showJSON, _ := cmd.Flags().GetBool("json")

		mgr, err := newManager()
		if err != nil {
			return err
		}

		stale, err := mgr.StaleWorkspaces(cmd.Context(), force)
		if err != nil {
			return err
		}

		if showJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(stale)
		}

		if !remove {
			if len(stale) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no stale workspaces found")
				return nil
			}
			for _, st := range stale {
				fmt.Fprintf(cmd.OutOrStdout(), "%s (branch: %s)\n", st.ID, st.Workspace.Branch)
			}
			return nil
		}

		var errs []error
		for _, st := range stale {
			if err := mgr.RemoveWorkspace(cmd.Context(), st.ID, workspace.RemoveOptions{Force: force}); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", st.ID, err))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", st.ID)
			}
		}

		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(staleCmd)
	staleCmd.Flags().Bool("json", false, "Output as JSON")
	staleCmd.Flags().Bool("rm", false, "Remove all stale workspaces (interactive)")
	staleCmd.Flags().Bool("force", false, "Force removal without confirmation")
}

package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all workspaces",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := newManager()
		if err != nil {
			return err
		}

		showAll, _ := cmd.Flags().GetBool("all")
		showJSON, _ := cmd.Flags().GetBool("json")
		repoFilter, _ := cmd.Flags().GetString("repo")

		statuses, err := mgr.ListWorkspaces(cmd.Context())
		if err != nil {
			return err
		}

		if repoFilter != "" {
			filtered := statuses[:0]
			for _, st := range statuses {
				if st.Workspace.Repo == repoFilter {
					filtered = append(filtered, st)
				}
			}
			statuses = filtered
		}

		if showJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(statuses)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		if showAll {
			fmt.Fprintln(w, "WORKSPACE\tSTATUS\tLAST ACCESSED\tWORKTREE\tBRANCH")
		} else {
			fmt.Fprintln(w, "WORKSPACE\tSTATUS\tLAST ACCESSED")
		}

		for _, st := range statuses {
			status := "dead"
			if st.SessionAlive {
				status = "alive"
			}
			last := st.Workspace.LastAccessedAt.Format(time.RFC3339)
			if showAll {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", st.ID, status, last, st.Workspace.WorktreePath, st.Workspace.Branch)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\n", st.ID, status, last)
			}
		}

		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().BoolP("all", "a", false, "Show all details")
	lsCmd.Flags().Bool("json", false, "Output as JSON")
	lsCmd.Flags().String("repo", "", "Filter by repository")
}

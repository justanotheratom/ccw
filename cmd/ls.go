package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/ccw/ccw/internal/workspace"
	"github.com/fatih/color"
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

		// Track global indices before filtering (1-based, matching lookupWorkspace)
		type indexedStatus struct {
			Index  int
			Status workspace.WorkspaceStatus
		}
		indexed := make([]indexedStatus, 0, len(statuses))
		for i, st := range statuses {
			if repoFilter == "" || st.Workspace.Repo == repoFilter {
				indexed = append(indexed, indexedStatus{Index: i + 1, Status: st})
			}
		}

		if showJSON {
			// For JSON, just output the filtered statuses (without index wrapper)
			filtered := make([]workspace.WorkspaceStatus, len(indexed))
			for i, is := range indexed {
				filtered[i] = is.Status
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(filtered)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		if showAll {
			fmt.Fprintln(w, "#\tWORKSPACE\tSTATUS\tLAST ACCESSED\tWORKTREE\tBRANCH")
		} else {
			fmt.Fprintln(w, "#\tWORKSPACE\tSTATUS\tLAST ACCESSED")
		}

		for _, is := range indexed {
			st := is.Status
			status := "dead"
			if st.SessionAlive {
				status = "alive"
			}
			coloredStatus := status
			if st.SessionAlive {
				coloredStatus = color.New(color.FgGreen).Sprint(status)
			} else {
				coloredStatus = color.New(color.FgRed).Sprint(status)
			}
			last := st.Workspace.LastAccessedAt.Format(time.RFC3339)
			if showAll {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", is.Index, st.ID, coloredStatus, last, st.Workspace.WorktreePath, st.Workspace.Branch)
			} else {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", is.Index, st.ID, coloredStatus, last)
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

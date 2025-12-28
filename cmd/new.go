package cmd

import (
	"fmt"
	"strings"

	"github.com/ccw/ccw/internal/deps"
	"github.com/ccw/ccw/internal/workspace"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <repo> <branch>",
	Short: "Create a new workspace",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := args[0]
		branch := args[1]

		base, _ := cmd.Flags().GetString("base")
		noAttach, _ := cmd.Flags().GetBool("no-attach")
		noFetch, _ := cmd.Flags().GetBool("no-fetch")
		message, _ := cmd.Flags().GetString("message")

		mgr, err := newManager()
		if err != nil {
			return err
		}

		warnOptionalDeps(cmd)

		ws, err := mgr.CreateWorkspace(cmd.Context(), repo, branch, workspace.CreateOptions{
			BaseBranch: base,
			NoAttach:   noAttach,
			NoFetch:    noFetch,
			Message:    message,
		})
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "created workspace %s at %s\n", workspace.WorkspaceID(repo, branch), ws.WorktreePath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringP("base", "b", "", "Base branch to create from (default: main)")
	newCmd.Flags().Bool("no-attach", false, "Create but don't attach to session")
	newCmd.Flags().StringP("message", "m", "", "Initial prompt to send to Claude Code")
	newCmd.Flags().Bool("no-fetch", false, "Skip fetch/prune of base (not recommended)")
}

func warnOptionalDeps(cmd *cobra.Command) {
	var hints []string
	for _, dep := range deps.DefaultDependencies() {
		if !dep.Optional {
			continue
		}
		res := deps.Check(dep)
		if !res.Found {
			name := dep.DisplayName
			if name == "" {
				name = dep.Name
			}
			hints = append(hints, fmt.Sprintf("%s: %s", name, dep.InstallHint))
		}
	}

	if len(hints) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "warning: optional dependencies missing: %s\n", strings.Join(hints, "; "))
	}
}

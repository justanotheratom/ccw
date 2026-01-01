package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ccw/ccw/internal/workspace"
	"github.com/fatih/color"
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
		yes, _ := cmd.Flags().GetBool("yes")

		mgr, err := newManager()
		if err != nil {
			return err
		}

		// Build confirmation function for interactive prompts
		var confirmFunc func(message string, files []string) bool
		if !force && !yes {
			confirmFunc = func(message string, files []string) bool {
				yellow := color.New(color.FgYellow)
				yellow.Fprintf(cmd.OutOrStdout(), "Warning: %s\n", message)

				if len(files) > 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "Files that differ:")
					for _, f := range files {
						fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", f)
					}
				}

				fmt.Fprint(cmd.OutOrStdout(), "Delete anyway? [y/N] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				return answer == "y" || answer == "yes"
			}
		}

		err = mgr.RemoveWorkspace(cmd.Context(), id, workspace.RemoveOptions{
			Force:        force,
			KeepBranch:   keepBranch,
			KeepWorktree: keepWorktree,
			ConfirmFunc:  confirmFunc,
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
	rmCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts (auto-confirm)")
}

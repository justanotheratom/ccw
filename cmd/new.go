package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <repo> <branch>",
	Short: "Create a new workspace",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("new: not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringP("base", "b", "", "Base branch to create from (default: main)")
	newCmd.Flags().Bool("no-attach", false, "Create but don't attach to session")
	newCmd.Flags().StringP("message", "m", "", "Initial prompt to send to Claude Code")
	newCmd.Flags().Bool("no-fetch", false, "Skip fetch/prune of base (not recommended)")
}

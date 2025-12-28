package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all workspaces",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("ls: not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().BoolP("all", "a", false, "Show all details")
	lsCmd.Flags().Bool("json", false, "Output as JSON")
	lsCmd.Flags().String("repo", "", "Filter by repository")
}

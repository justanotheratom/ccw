package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var staleCmd = &cobra.Command{
	Use:   "stale",
	Short: "List workspaces with merged branches",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("stale: not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(staleCmd)
	staleCmd.Flags().Bool("rm", false, "Remove all stale workspaces (interactive)")
	staleCmd.Flags().Bool("force", false, "Force removal without confirmation")
}

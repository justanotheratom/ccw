package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <workspace>",
	Short: "Open or attach to an existing workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("open: not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
	openCmd.Flags().Bool("no-resume", false, "Do not resume Claude Code session")
}

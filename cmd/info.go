package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <workspace>",
	Short: "Show detailed information about a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("info: not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

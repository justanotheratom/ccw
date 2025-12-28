package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "View or edit configuration",
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("config: not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().Bool("reset", false, "Reset configuration to defaults")
}

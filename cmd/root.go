package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.5"
)

func Execute() error {
	rootCmd.SilenceUsage = true
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "ccw",
	Short: "Claude Code Workspace manager",
	Long:  "ccw is a CLI tool for managing Claude Code workspaces with git worktrees and tmux sessions.",
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print ccw version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stdout, version)
	},
}

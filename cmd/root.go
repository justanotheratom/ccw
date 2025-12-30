package cmd

import (
	"fmt"
	"os"

	"github.com/ccw/ccw/internal/config"
	"github.com/ccw/ccw/internal/onboarding"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.59"

	// Commands exempt from onboarding requirement
	onboardingExemptCmds = map[string]bool{
		"version":    true,
		"help":       true,
		"completion": true,
	}
)

func Execute() error {
	rootCmd.SilenceUsage = true
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "ccw",
	Short: "Claude Code Workspace manager",
	Long:  "ccw is a CLI tool for managing Claude Code workspaces with git worktrees and tmux sessions.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip onboarding for exempt commands
		if onboardingExemptCmds[cmd.Name()] {
			return nil
		}

		// Check if onboarding is needed
		store, err := config.NewStore("")
		if err != nil {
			return err
		}

		cfg, err := store.Load()
		if err != nil {
			return err
		}

		if onboarding.NeedsOnboarding(cfg) {
			o := onboarding.New(store)
			if _, err := o.Run(); err != nil {
				return fmt.Errorf("onboarding failed: %w", err)
			}
		}

		return nil
	},
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

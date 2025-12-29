package cmd

import (
	"fmt"
	"strings"

	"github.com/ccw/ccw/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "View or edit configuration",
	Args:  cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		reset, _ := cmd.Flags().GetBool("reset")

		mgr, err := newManager()
		if err != nil {
			return err
		}

		if reset {
			cfg, err := mgr.ResetConfig()
			if err != nil {
				return err
			}
			printConfig(cmd, cfg)
			return nil
		}

		switch len(args) {
		case 0:
			printConfig(cmd, mgr.GetConfig())
		case 1:
			value, err := configValue(mgr.GetConfig(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
		case 2:
			cfg, err := mgr.SetConfigValue(args[0], args[1])
			if err != nil {
				return err
			}
			printConfig(cmd, cfg)
		default:
			return fmt.Errorf("too many arguments")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().Bool("reset", false, "Reset configuration to defaults")
}

func printConfig(cmd *cobra.Command, cfg config.Config) {
	builder := []string{
		fmt.Sprintf("repos_dir=%s", cfg.ReposDir),
		fmt.Sprintf("iterm_cc_mode=%t", cfg.ITermCCMode),
		fmt.Sprintf("claude_rename_delay=%d", cfg.ClaudeRenameDelay),
		fmt.Sprintf("layout.left=%s", cfg.Layout.Left),
		fmt.Sprintf("layout.right=%s", cfg.Layout.Right),
		fmt.Sprintf("claude_dangerously_skip_permissions=%t", cfg.ClaudeDangerouslySkipPerms),
		fmt.Sprintf("onboarded=%t", cfg.Onboarded),
	}
	fmt.Fprintln(cmd.OutOrStdout(), strings.Join(builder, "\n"))
}

func configValue(cfg config.Config, key string) (string, error) {
	switch key {
	case "repos_dir":
		return cfg.ReposDir, nil
	case "iterm_cc_mode":
		return fmt.Sprintf("%t", cfg.ITermCCMode), nil
	case "claude_rename_delay":
		return fmt.Sprintf("%d", cfg.ClaudeRenameDelay), nil
	case "layout.left":
		return cfg.Layout.Left, nil
	case "layout.right":
		return cfg.Layout.Right, nil
	case "claude_dangerously_skip_permissions":
		return fmt.Sprintf("%t", cfg.ClaudeDangerouslySkipPerms), nil
	case "onboarded":
		return fmt.Sprintf("%t", cfg.Onboarded), nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}

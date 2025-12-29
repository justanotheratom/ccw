package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var reposCmd = &cobra.Command{
	Use:   "repos",
	Short: "List available repositories",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := newManager()
		if err != nil {
			return err
		}

		cfg := mgr.GetConfig()
		reposDir, err := cfg.ExpandedReposDir()
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(reposDir)
		if err != nil {
			return err
		}

		var repos []string
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				repos = append(repos, e.Name())
			}
		}
		sort.Strings(repos)

		showJSON, _ := cmd.Flags().GetBool("json")
		if showJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(repos)
		}
		for _, r := range repos {
			fmt.Fprintln(cmd.OutOrStdout(), r)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reposCmd)
	reposCmd.Flags().Bool("json", false, "Output as JSON")
}

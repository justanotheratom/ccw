package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/ccw/ccw/internal/deps"
	"github.com/spf13/cobra"
)

type DepStatus struct {
	Installed bool   `json:"installed"`
	Path      string `json:"path"`
	Optional  bool   `json:"optional,omitempty"`
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check dependencies",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		result := make(map[string]DepStatus)
		for _, d := range deps.DefaultDependencies() {
			res := deps.Check(d)
			result[d.Name] = DepStatus{
				Installed: res.Found,
				Path:      res.Path,
				Optional:  d.Optional,
			}
		}

		showJSON, _ := cmd.Flags().GetBool("json")
		if showJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		}
		for name, st := range result {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%t\t%s\n", name, st.Installed, st.Path)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().Bool("json", false, "Output as JSON")
}

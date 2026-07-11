package cli

import "github.com/spf13/cobra"

var diffCmd = &cobra.Command{
	Use:   "diff <old.json> <new.json>",
	Short: "Diff two snapshots to show topology change over time",
	Args:  cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, _ []string) error {
		return errNotImplemented("diff")
	},
}

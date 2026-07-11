package cli

import "github.com/spf13/cobra"

var (
	explainFrom string
	explainTo   string
	explainPort int
)

var explainCmd = &cobra.Command{
	Use:   "explain <topology.json>",
	Short: "Explain the reachability path between two resources",
	Long: "explain reasons over a snapshot (no AWS) to report the path from one resource\n" +
		"to another, the first blocking hop when a path is broken, and an honest verdict\n" +
		"of what was and was not evaluated.",
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, _ []string) error {
		return errNotImplemented("explain")
	},
}

func init() {
	f := explainCmd.Flags()
	f.StringVar(&explainFrom, "from", "", "source resource id (required)")
	f.StringVar(&explainTo, "to", "", "destination resource id (required)")
	f.IntVar(&explainPort, "port", 0, "destination port")
	_ = explainCmd.MarkFlagRequired("from")
	_ = explainCmd.MarkFlagRequired("to")
}

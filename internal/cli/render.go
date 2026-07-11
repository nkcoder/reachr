package cli

import "github.com/spf13/cobra"

var (
	renderFormat string
	renderOutput string
)

var renderCmd = &cobra.Command{
	Use:   "render <topology.json>",
	Short: "Render a snapshot to an interactive diagram (pure function, no AWS)",
	Long: "render is a pure function of a topology.json snapshot: it makes no AWS calls.\n" +
		"Default output is a self-contained interactive HTML diagram; --format dot is\n" +
		"the Graphviz escape hatch.",
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, _ []string) error {
		return errNotImplemented("render")
	},
}

func init() {
	f := renderCmd.Flags()
	f.StringVar(&renderFormat, "format", "html", "output format: html|dot")
	f.StringVarP(&renderOutput, "output", "o", "topology.html", "output path")
}

package cli

import "github.com/spf13/cobra"

var (
	scanRegion  string
	scanProfile string
	scanFilter  string
	scanOutput  string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan AWS (read-only) into an immutable topology.json snapshot",
	Long: "scan reads the target account/region via the AWS default credential chain and\n" +
		"writes a topology.json snapshot. It is the only phase that talks to AWS.",
	RunE: func(_ *cobra.Command, _ []string) error {
		return errNotImplemented("scan")
	},
}

func init() {
	f := scanCmd.Flags()
	f.StringVar(&scanRegion, "region", "", "AWS region to scan (required)")
	f.StringVar(&scanProfile, "profile", "", "AWS profile (default credential chain if empty)")
	f.StringVar(&scanFilter, "filter", "", "scope selector, e.g. tag:project=X")
	f.StringVarP(&scanOutput, "output", "o", "topology.json", "snapshot output path")
	_ = scanCmd.MarkFlagRequired("region")
}

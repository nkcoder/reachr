package cli

import (
	"encoding/json"
	"fmt"

	"github.com/nkcoder/reachr/internal/awsscan"
	"github.com/spf13/cobra"
)

var (
	scanRegion  string
	scanProfile string
	scanFilter  string
	scanOutput  string
	scanRaw     bool
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan AWS (read-only) into an immutable topology.json snapshot",
	Long: "scan reads the target account/region via the AWS default credential chain and\n" +
		"writes a topology.json snapshot. It is the only phase that talks to AWS.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if !scanRaw {
			return errNotImplemented("scan")
		}
		return runRawDump(cmd)
	},
}

// runRawDump is the exploratory backbone dump (issue #6): it prints raw EC2
// Describe* output so we can design the topology.json schema from real shapes.
func runRawDump(cmd *cobra.Command) error {
	ctx := cmd.Context()
	client, err := awsscan.NewClient(ctx, scanRegion, scanProfile)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}
	raw, err := client.DumpBackbone(ctx)
	if err != nil {
		return fmt.Errorf("dump backbone: %w", err)
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(raw)
}

func init() {
	f := scanCmd.Flags()
	f.StringVar(&scanRegion, "region", "", "AWS region to scan (required)")
	f.StringVar(&scanProfile, "profile", "", "AWS profile (default credential chain if empty)")
	f.StringVar(&scanFilter, "filter", "", "scope selector, e.g. tag:project=X")
	f.StringVarP(&scanOutput, "output", "o", "topology.json", "snapshot output path")
	f.BoolVar(&scanRaw, "raw", false, "exploratory: dump raw AWS backbone JSON to stdout")
	_ = scanCmd.MarkFlagRequired("region")
}

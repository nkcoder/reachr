package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nkcoder/reachr/internal/awsscan"
	"github.com/nkcoder/reachr/internal/topology"
	"github.com/spf13/cobra"
)

var (
	scanRegion  string
	scanProfile string
	scanVPC     string
	scanFilter  string
	scanOutput  string
	scanRaw     bool
	scanScrub   bool
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan AWS (read-only) into an immutable topology.json snapshot",
	Long: "scan reads the target account/region via the AWS default credential chain and\n" +
		"writes a topology.json snapshot. It is the only phase that talks to AWS.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		if scanRaw {
			return runRawDump(cmd)
		}
		return runScan(cmd)
	},
}

// runScan collects the backbone and writes a normalized topology.json snapshot.
func runScan(cmd *cobra.Command) error {
	ctx := cmd.Context()
	client, err := awsscan.NewClient(ctx, scanRegion, scanProfile)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}
	account, err := client.AccountID(ctx)
	if err != nil {
		return fmt.Errorf("resolve account: %w", err)
	}
	raw, err := client.DumpBackbone(ctx, scanVPC)
	if err != nil {
		return fmt.Errorf("scan backbone: %w", err)
	}
	snap := raw.ToSnapshot(account, scanFilter)
	if scanScrub {
		topology.Sanitize(snap)
	}

	f, err := os.Create(scanOutput)
	if err != nil {
		return fmt.Errorf("create %s: %w", scanOutput, err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		_ = f.Close()
		return fmt.Errorf("write snapshot: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", scanOutput, err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "wrote %s (%d VPCs, %d ENIs, %d SGs)\n",
		scanOutput, len(snap.VPCs), len(snap.ENIs), len(snap.SecurityGroups))
	return nil
}

// runRawDump is the exploratory backbone dump (issue #6): it prints raw EC2
// Describe* output so we can design the topology.json schema from real shapes.
func runRawDump(cmd *cobra.Command) error {
	ctx := cmd.Context()
	client, err := awsscan.NewClient(ctx, scanRegion, scanProfile)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}
	raw, err := client.DumpBackbone(ctx, scanVPC)
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
	f.StringVar(&scanVPC, "vpc", "", "scope scan to a single VPC id (default: whole region)")
	f.StringVar(&scanFilter, "filter", "", "scope selector, e.g. tag:project=X")
	f.StringVarP(&scanOutput, "output", "o", "topology.json", "snapshot output path")
	f.BoolVar(&scanRaw, "raw", false, "exploratory: dump raw AWS backbone JSON to stdout")
	f.BoolVar(&scanScrub, "scrub", false, "redact the AWS account id (for shareable snapshots / fixtures)")
	_ = scanCmd.MarkFlagRequired("region")
}

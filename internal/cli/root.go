// Package cli defines the reachr command tree (scan/render/explain/diff).
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "reachr",
	Short: "AWS network reachability & topology tool",
	Long: "reachr scans AWS (read-only) into an immutable topology.json snapshot, then\n" +
		"renders an interactive whole-VPC topology and explains what can actually\n" +
		"reach what — and why a given path is broken.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(scanCmd, renderCmd, explainCmd, diffCmd)
}

// errNotImplemented is the placeholder verdict for stubbed subcommands.
func errNotImplemented(name string) error {
	return fmt.Errorf("%s: not implemented yet", name)
}

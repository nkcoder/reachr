package cli

import "testing"

func TestRootHasExpectedSubcommands(t *testing.T) {
	got := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		got[c.Name()] = true
	}
	for _, want := range []string{"scan", "render", "explain", "diff"} {
		if !got[want] {
			t.Errorf("root command missing subcommand %q", want)
		}
	}
}

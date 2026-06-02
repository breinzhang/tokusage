package cli

import "github.com/spf13/cobra"

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "tokusage",
		Short:         "Analyze local agent token usage",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newClaudeCommand())
	cmd.AddCommand(newCacheCommand())
	return cmd
}

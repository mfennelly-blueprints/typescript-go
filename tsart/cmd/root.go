package cmd

import (
	"fmt"

	"github.com/microsoft/typescript-go/internal/core"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "tsart",
	Short:   "TypeScript artifact tooling",
	Version: core.Version(),
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the tsart version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Fprintln(cmd.OutOrStdout(), cmd.Root().Name()+" version "+cmd.Root().Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the tsart command-line interface.
func Execute() error {
	return rootCmd.Execute()
}

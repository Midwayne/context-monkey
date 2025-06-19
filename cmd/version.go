package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// This will be set by the linker
var version string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of context-monkey",
	Run: func(cmd *cobra.Command, args []string) {
		if version == "" {
			// fallback for local builds
			version = "dev"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "context-monkey version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

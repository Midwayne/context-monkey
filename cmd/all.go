package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Variables to store flag values for the all command
var (
	allOutputDir  string
	allIgnore     []string // Using string slice for multiple ignore patterns
	allProjectDir string
)

// allCmd represents the all command
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Concatenate all relevant project files (e.g., code, configs)",
	Long: `The 'all' command processes the specified project directory,
		respecting .gitignore and custom ignore patterns and concatenates
		the content of all relevant files and project structure into a single output.
		This is useful for creating a comprehensive context snapshot of your project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Building project context...")
		fmt.Printf("  Project Directory: %s\n", allProjectDir)
		fmt.Printf("  Output File: %s\n", allOutputDir)
		fmt.Printf("  Ignore Patterns: %v\n", allIgnore)

		// TODO: Implement the logic for the 'all' command:
		// 1. Resolve project directory.
		// 2. List files using 'git ls-files' (or similar to respect .gitignore).
		// 3. Filter files based on 'allIgnore' patterns.
		// 4. Read content of each remaining file.
		// 5. Concatenate content with separators (ex: "--- START FILE: path/to/file ---").
		// 6. Write to 'allOutputDir' or stdout if empty.

		if allProjectDir == "" {
			// Default to current directory if not set by persistent flag or local flag
			allProjectDir = "."
		}

		fmt.Println("'all' command logic to be implemented.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(allCmd)

	// Local flags for the 'all' command
	allCmd.Flags().StringVarP(&allProjectDir, "dir", "d", ".", "Path to the project directory")
	allCmd.Flags().StringVarP(&allOutputDir, "output", "o", "", "Output file path for the concatenated content (default: stdout)")
	allCmd.Flags().StringSliceVarP(&allIgnore, "ignore", "i", []string{}, "Comma-separated glob patterns of files/directories to ignore (e.g., 'tests/*,*.log')")

	// Example of marking a flag as required:
	// allCmd.MarkFlagRequired("dir")
}

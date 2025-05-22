package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Variables to store flag values for the files command
var (
	filesOutputDir  string
	filesIgnore     []string
	filesProjectDir string
)

// filesCmd represents the files command
var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Concatenate specified project files",
	Long: `The 'files' command concatenates the content of project files.
		You can specify which files to include or exclude using ignore patterns.
		This command is useful for gathering specific code or text parts for an LLM.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Executing 'files' command...")
		fmt.Printf("  Project Directory: %s\n", filesProjectDir)
		fmt.Printf("  Output File: %s\n", filesOutputDir)
		fmt.Printf("  Ignore Patterns: %v\n", filesIgnore)

		// TODO: Implement the logic for the 'files' command:
		// Similar to the all command

		if filesProjectDir == "" {
			filesProjectDir = "."
		}

		fmt.Println("'files' command logic to be implemented.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(filesCmd)

	filesCmd.Flags().StringVarP(&filesProjectDir, "dir", "d", ".", "Path to the project directory")
	filesCmd.Flags().StringVarP(&filesOutputDir, "output", "o", "", "Output file path for the concatenated files (default: stdout)")
	filesCmd.Flags().StringSliceVarP(&filesIgnore, "ignore", "i", []string{}, "Comma-separated glob patterns of files/directories to ignore")
}

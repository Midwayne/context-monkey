package cmd

import (
	"como/utils"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	treeOutputDir  string
	treeIgnore     []string
	treeProjectDir string
)

// treeCmd represents the tree command
var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Generate a file tree listing of the project",
	Long: `The 'tree' command lists files and directories in the project directory,
			respecting .gitignore and custom ignore patterns.
			It outputs a structured tree representation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "Executing 'tree' command...")

		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		if treeProjectDir == "" || treeProjectDir == "." {
			treeProjectDir = currentDir
		} else {
			treeProjectDir, err = filepath.Abs(treeProjectDir)
			if err != nil {
				return fmt.Errorf("failed to resolve project directory path %s: %w", treeProjectDir, err)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  Project Directory: %s\n", treeProjectDir)
		if treeOutputDir != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Output File: %s\n", treeOutputDir)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "  Output: stdout")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  Ignore Patterns: %v\n", treeIgnore)

		// 1. Generate file tree string
		// TODO: Design consideration - include ignored files or not?
		treeString, err := utils.BuildFileTree(treeProjectDir, treeIgnore, true, nil, true)

		if err != nil {
			return fmt.Errorf("failed to generate file tree: %w", err)
		}

		// 2. Get output writer
		writer, outFile, err := utils.GetOutputWriter(treeOutputDir)
		if err != nil {
			return err
		}
		if outFile != nil {
			defer outFile.Close()
			defer writer.Flush()
		} else {
			defer writer.Flush() // Flush for stdout
		}

		// 3. Write tree string to output
		if _, err := writer.WriteString(treeString); err != nil {
			return fmt.Errorf("failed to write tree to output: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "File tree generated successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(treeCmd)

	treeCmd.Flags().StringVarP(&treeProjectDir, "dir", "d", ".", "Path to the project directory")
	treeCmd.Flags().StringVarP(&treeOutputDir, "output", "o", "", "Output file path for the file tree (default: stdout, use '-' for stdout)")
	treeCmd.Flags().StringSliceVarP(&treeIgnore, "ignore", "i", []string{}, "Comma-separated glob patterns of files/directories to ignore")
}

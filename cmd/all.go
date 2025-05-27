package cmd

import (
	"como/utils"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	allOutputDir  string
	allIgnore     []string
	allProjectDir string
	allSkipBinary bool
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
		fmt.Fprintln(cmd.OutOrStdout(), "Building project context...")

		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		// Resolve project directory
		if allProjectDir == "" || allProjectDir == "." {
			allProjectDir = currentDir
		} else {
			allProjectDir, err = filepath.Abs(allProjectDir)
			if err != nil {
				return fmt.Errorf("failed to resolve project directory path %s: %w", allProjectDir, err)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  Project Directory: %s\n", allProjectDir)
		if allOutputDir != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Output File: %s\n", allOutputDir)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "  Output: stdout")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  Ignore Patterns: %v\n", allIgnore)
		fmt.Fprintf(cmd.OutOrStdout(), "  Skip Binary Files: %v\n", allSkipBinary)

		// 1. List files
		// For 'all' command, specificFileArgs is nil as we scan the directory.
		// We don't include directories in the result for concatenation.
		filesToProcess, err := utils.GetProjectFiles(allProjectDir, allIgnore, true, nil, false)
		if err != nil {
			return fmt.Errorf("failed to list project files: %w", err)
		}

		if len(filesToProcess) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No files found to process after applying ignores.")
			return nil
		}

		// 2. Get output writer
		writer, outFile, err := utils.GetOutputWriter(allOutputDir)
		if err != nil {
			return err
		}
		if outFile != nil {
			defer outFile.Close()
			// Ensure buffered content is written before file close
			defer writer.Flush()
		} else {
			// Flush for stdout as well
			defer writer.Flush()
		}

		// 3. Generate and write tree
		treeString, err := utils.BuildFileTree(allProjectDir, allIgnore, true, nil, true)
		if err != nil {
			return fmt.Errorf("failed to generate file tree: %w", err)
		}

		if _, err := writer.WriteString("--- START FILE: PROJECT STRUCTURE ---\n"); err != nil {
			return fmt.Errorf("failed to write tree start marker: %w", err)
		}
		if _, err := writer.WriteString(treeString); err != nil {
			return fmt.Errorf("failed to write tree content: %w", err)
		}
		if _, err := writer.WriteString("\n--- END FILE: PROJECT STRUCTURE ---\n\n"); err != nil {
			return fmt.Errorf("failed to write tree end marker: %w", err)
		}

		// 4. Read content of each remaining file and concatenate
		fmt.Fprintln(cmd.OutOrStdout(), "Concatenating files...")
		for _, fileInfo := range filesToProcess {
			// Skip directories and symlinks for concatenation
			if fileInfo.IsDir || fileInfo.IsSymlink {
				// GetProjectFiles with includeDirsInResult=false should already filter these, but double check.
				// If it's a symlink to a file, it might be processed if not filtered earlier.
				// For now, GetProjectFiles with false should only return regular files.
				continue
			}

			content, isBinary, err := utils.ReadFileContent(fileInfo.AbsPath)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: skipping file %s due to read error: %v\n", fileInfo.RelPath, err)
				continue
			}

			if isBinary && allSkipBinary {
				fmt.Fprintf(cmd.OutOrStdout(), "  Skipping binary file: %s\n", fileInfo.RelPath)
				continue
			}

			// Write separator and content
			separatorStart := fmt.Sprintf("--- START FILE: %s ---\n", fileInfo.RelPath)
			separatorEnd := fmt.Sprintf("\n--- END FILE: %s ---\n\n", fileInfo.RelPath)

			if _, err := writer.WriteString(separatorStart); err != nil {
				return fmt.Errorf("failed to write start separator for %s: %w", fileInfo.RelPath, err)
			}
			if _, err := writer.WriteString(content); err != nil {
				return fmt.Errorf("failed to write content for %s: %w", fileInfo.RelPath, err)
			}
			if _, err := writer.WriteString(separatorEnd); err != nil {
				return fmt.Errorf("failed to write end separator for %s: %w", fileInfo.RelPath, err)
			}
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Project context built successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(allCmd)

	allCmd.Flags().StringVarP(&allProjectDir, "dir", "d", ".", "Path to the project directory")
	allCmd.Flags().StringVarP(&allOutputDir, "output", "o", "", "Output file path for the concatenated content (default: stdout, use '-' for stdout)")
	allCmd.Flags().StringSliceVarP(&allIgnore, "ignore", "i", []string{}, "Comma-separated glob patterns of files/directories to ignore (e.g., 'tests/*,*.log')")
	allCmd.Flags().BoolVar(&allSkipBinary, "skip-binary", true, "Skip binary files from concatenation")
}

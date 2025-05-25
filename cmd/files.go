package cmd

import (
	"como/utils"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	filesOutputDir  string
	filesIgnore     []string
	filesProjectDir string
	filesSkipBinary bool
)

// filesCmd represents the files command
var filesCmd = &cobra.Command{
	Use:   "files [file_or_glob1] [file_or_glob2]...",
	Short: "Concatenate specified project files or glob patterns",
	Long: `The 'files' command concatenates the content of specified project files or files matching glob patterns.
		You can specify which files to include via arguments and further exclude using ignore patterns.
		The --dir flag acts as the base directory for resolving relative file paths and glob patterns.
		This command is useful for gathering specific code or text parts for an LLM.`,
	Args: cobra.MinimumNArgs(1), // Require at least one file/glob argument
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "Executing 'files' command...")

		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		// Resolve project directory (base for relative paths in args)
		if filesProjectDir == "" || filesProjectDir == "." {
			filesProjectDir = currentDir
		} else {
			filesProjectDir, err = filepath.Abs(filesProjectDir)
			if err != nil {
				return fmt.Errorf("failed to resolve project directory path %s: %w", filesProjectDir, err)
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  Project Directory (base for files): %s\n", filesProjectDir)
		fmt.Fprintf(cmd.OutOrStdout(), "  Files/Globs to process: %v\n", args)
		if filesOutputDir != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Output File: %s\n", filesOutputDir)
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "  Output: stdout")
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  Ignore Patterns: %v\n", filesIgnore)
		fmt.Fprintf(cmd.OutOrStdout(), "  Skip Binary Files: %v\n", filesSkipBinary)

		// 1. List files based on arguments and apply ignores
		// For 'files' command, specificFileArgs is args from CLI.
		// We don't include directories in the result for concatenation.
		filesToProcess, err := utils.GetProjectFiles(filesProjectDir, filesIgnore, true, args, false)
		if err != nil {
			return fmt.Errorf("failed to list specified project files: %w", err)
		}

		if len(filesToProcess) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No files found matching the arguments after applying ignores.")
			return nil
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Files to be concatenated:")
		for _, fi := range filesToProcess {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", fi.RelPath)
		}

		// 2. Get output writer
		writer, outFile, err := utils.GetOutputWriter(filesOutputDir)
		if err != nil {
			return err
		}
		if outFile != nil {
			defer outFile.Close()
			defer writer.Flush()
		} else {
			defer writer.Flush()
		}

		// 3. Read content of each remaining file and concatenate
		fmt.Fprintln(cmd.OutOrStdout(), "Concatenating files...")
		for _, fileInfo := range filesToProcess {
			// Should be filtered by GetProjectFiles with includeDirsInResult=false
			if fileInfo.IsDir || fileInfo.IsSymlink {
				continue
			}

			content, isBinary, err := utils.ReadFileContent(fileInfo.AbsPath)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: skipping file %s due to read error: %v\n", fileInfo.RelPath, err)
				continue
			}

			if isBinary && filesSkipBinary {
				fmt.Fprintf(cmd.OutOrStdout(), "  Skipping binary file: %s\n", fileInfo.RelPath)
				continue
			}

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

		fmt.Fprintln(cmd.OutOrStdout(), "'files' command executed successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(filesCmd)

	filesCmd.Flags().StringVarP(&filesProjectDir, "dir", "d", ".", "Base path for resolving file arguments and .gitignore")
	filesCmd.Flags().StringVarP(&filesOutputDir, "output", "o", "", "Output file path for the concatenated files (default: stdout, use '-' for stdout)")
	filesCmd.Flags().StringSliceVarP(&filesIgnore, "ignore", "i", []string{}, "Comma-separated glob patterns of files/directories to ignore from the specified list")
	filesCmd.Flags().BoolVar(&filesSkipBinary, "skip-binary", true, "Skip binary files from concatenation")
}

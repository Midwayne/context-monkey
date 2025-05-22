package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "context-monkey",
	Aliases: []string{"como"},
	Short:   "A CLI tool to generate project context for LLMs",
	Long: `context-monkey is a CLI application that helps you prepare project 
			files and structures for use as context with Large Language Models.
			You can concatenate files, generate file trees and more, with
			options to ignore specific files or directories and specify output locations.`,
	Version: "0.0.1",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main. It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// Common flags that might be used across multiple commands can be defined here
// as persistent flags on the rootCmd.
// var projectDir string // Example: if --dir was global

func init() {
	// cobra.OnInitialize(initConfig)

	// Example of a persistent flag available to all subcommands:
	// rootCmd.PersistentFlags().StringVar(&projectDir, "dir", ".", "Path to the project directory")

	// Subcommands go here. They will be defined in their own files.
	// e.g., rootCmd.AddCommand(allCmd)
	//       rootCmd.AddCommand(filesCmd)
	//       rootCmd.AddCommand(treeCmd)
	//       rootCmd.AddCommand(updateCmd)
}

// initConfig reads in config file and ENV variables if set.
// func initConfig() {
// 	// ...
// }

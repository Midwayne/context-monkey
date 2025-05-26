package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// ReadFileContent reads the content of a file into a string.
func ReadFileContent(filePath string) (string, bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	checkLen := 1024
	if len(content) < checkLen {
		checkLen = len(content)
	}
	if bytes.Contains(content[:checkLen], []byte{0}) {
		return "", true, nil
	}

	return string(content), false, nil
}

// GetOutputWriter returns a writer to the specified output file or os.Stdout.
func GetOutputWriter(outputDir string) (*bufio.Writer, *os.File, error) {
	if outputDir == "" || outputDir == "-" {
		return bufio.NewWriter(os.Stdout), nil, nil
	}

	dir := filepath.Dir(outputDir)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("failed to create output directory %s: %w", dir, err)
		}
	}

	file, err := os.Create(outputDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file %s: %w", outputDir, err)
	}
	return bufio.NewWriter(file), file, nil
}

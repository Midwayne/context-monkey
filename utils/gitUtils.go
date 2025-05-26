package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// isGitRepo checks if the given directory is part of a Git repository.
func isGitRepo(dir string) bool {
	gitPath := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return true
	}

	// Check if we are in a worktree or a subdirectory of a repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

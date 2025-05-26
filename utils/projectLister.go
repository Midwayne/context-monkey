package utils

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	gitignore "github.com/sabhiram/go-gitignore"
)

// GetProjectFiles lists files in a project directory.
// rootDir: The root directory of the project.
// customIgnorePatterns: Glob patterns for files/dirs to ignore.
// respectGitIgnore: Whether to respect .gitignore rules.
// specificFileArgs: If non-empty, these are specific files/globs to process (used by 'files' command).
// includeDirsInResult: Whether to include directories in the returned list (useful for tree building).
func GetProjectFiles(
	rootDir string, customIgnorePatterns []string, respectGitIgnore bool,
	specificFileArgs []string, includeDirsInResult bool) ([]FileInfo, error) {

	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for rootDir %s: %w", rootDir, err)
	}

	var gitIgnoreMatcher *gitignore.GitIgnore
	if respectGitIgnore {
		gitIgnoreFilePath := filepath.Join(absRootDir, ".gitignore")
		if _, err := os.Stat(gitIgnoreFilePath); err == nil {
			compiledMatcher, compileErr := gitignore.CompileIgnoreFile(gitIgnoreFilePath)
			if compileErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not compile .gitignore at %s: %v\n", gitIgnoreFilePath, compileErr)
			} else {
				gitIgnoreMatcher = compiledMatcher
			}
		}
	}

	customMatchers := make([]glob.Glob, 0, len(customIgnorePatterns))
	for _, pattern := range customIgnorePatterns {
		g, err := glob.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid custom ignore pattern %s: %w", pattern, err)
		}
		customMatchers = append(customMatchers, g)
	}

	candidateFiles := make(map[string]FileInfo)

	if len(specificFileArgs) > 0 {
		for _, arg := range specificFileArgs {
			patternToGlob := arg
			if !filepath.IsAbs(arg) {
				patternToGlob = filepath.Join(absRootDir, arg)
			}

			matches, err := filepath.Glob(patternToGlob)
			if err != nil {
				return nil, fmt.Errorf("error expanding glob pattern %s: %w", arg, err)
			}
			for _, matchPath := range matches {
				absMatchPath, err := filepath.Abs(matchPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not get absolute path for %s: %v\n", matchPath, err)
					continue
				}
				relPath, err := filepath.Rel(absRootDir, absMatchPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not get relative path for %s (base: %s): %v\n", absMatchPath, absRootDir, err)
					relPath = filepath.Base(absMatchPath)
				}

				fileInfo, err := os.Lstat(absMatchPath)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					fmt.Fprintf(os.Stderr, "Warning: could not stat file %s: %v\n", absMatchPath, err)
					continue
				}
				isDir := fileInfo.IsDir()
				isSymlink := fileInfo.Mode()&os.ModeSymlink != 0

				candidateFiles[absMatchPath] = FileInfo{AbsPath: absMatchPath, RelPath: relPath, IsDir: isDir, IsSymlink: isSymlink}
			}
		}
	} else {
		useGitLsFiles := isGitRepo(absRootDir)
		if useGitLsFiles {
			cmd := exec.Command("git", "ls-files", "-coz", "--exclude-standard", "--full-name", "--")
			cmd.Dir = absRootDir

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr,
					"Warning: 'git ls-files' failed in %s (falling back to filesystem walk): %v\nStderr: %s\n", absRootDir, err, stderr.String())
				useGitLsFiles = false
			} else {
				repoRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
				repoRootCmd.Dir = absRootDir
				repoRootOutput, repoRootErr := repoRootCmd.Output()
				var actualRepoRoot string
				if repoRootErr == nil {
					actualRepoRoot = strings.TrimSpace(string(repoRootOutput))
				} else {
					actualRepoRoot = absRootDir
					fmt.Fprintf(os.Stderr, "Warning: could not determine git repo root for %s, assuming it is the project directory: %v\n", absRootDir, repoRootErr)
				}

				files := strings.Split(strings.TrimRight(stdout.String(), "\x00"), "\x00")
				for _, pathInRepo := range files {
					if pathInRepo == "" {
						continue
					}
					absPath := filepath.Join(actualRepoRoot, pathInRepo)

					if !strings.HasPrefix(absPath, absRootDir+string(filepath.Separator)) && absPath != absRootDir {
						continue
					}

					relPathToProjectRoot, err := filepath.Rel(absRootDir, absPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: could not make path %s relative to %s: %v\n", absPath, absRootDir, err)
						continue
					}

					fileInfo, err := os.Lstat(absPath)
					if err != nil {
						if os.IsNotExist(err) {
							continue
						}
						fmt.Fprintf(os.Stderr, "Warning: could not stat file from git ls-files %s: %v\n", absPath, err)
						continue
					}
					isDir := fileInfo.IsDir()
					isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
					candidateFiles[absPath] = FileInfo{AbsPath: absPath, RelPath: relPathToProjectRoot, IsDir: isDir, IsSymlink: isSymlink}
				}
			}
		}

		if !useGitLsFiles {
			err := filepath.WalkDir(absRootDir, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: error accessing path %s: %v\n", path, walkErr)
					if d.IsDir() && path != absRootDir {
						return filepath.SkipDir
					}
					return nil
				}

				relPath, Rerr := filepath.Rel(absRootDir, path)
				if Rerr != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not get relative path for %s: %v\n", path, Rerr)
					relPath = filepath.Base(path)
				}

				if path == absRootDir || relPath == "." {
					return nil
				}

				if d.Name() == ".git" && d.IsDir() {
					return filepath.SkipDir
				}

				if gitIgnoreMatcher != nil {
					if gitIgnoreMatcher.MatchesPath(relPath) {
						if d.IsDir() {
							return filepath.SkipDir
						}
						return nil // Skip ignored file
					}
				}

				isDir := d.IsDir()
				isSymlink := d.Type()&os.ModeSymlink != 0
				candidateFiles[path] = FileInfo{AbsPath: path, RelPath: relPath, IsDir: isDir, IsSymlink: isSymlink}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("error walking directory %s: %w", absRootDir, err)
			}
		}
	}

	var result []FileInfo
	for _, fi := range candidateFiles {
		pathForMatching := fi.RelPath

		if gitIgnoreMatcher != nil {
			if gitIgnoreMatcher.MatchesPath(fi.RelPath) {
				continue
			}
		}

		isCustomIgnored := false
		for _, matcher := range customMatchers {
			pathToTestWithGlob := filepath.ToSlash(pathForMatching)
			if matcher.Match(pathToTestWithGlob) {
				isCustomIgnored = true
				break
			}
		}
		if isCustomIgnored {
			continue
		}

		if !fi.IsDir || includeDirsInResult {
			result = append(result, fi)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].RelPath < result[j].RelPath
	})

	return result, nil
}

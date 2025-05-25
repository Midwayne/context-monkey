package utils

import (
	"bufio"
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

// FileInfo holds path information for a file.
type FileInfo struct {
	AbsPath   string // Absolute path to the file
	RelPath   string // Path relative to the project root
	IsDir     bool   // True if it's a directory
	IsSymlink bool   // True if it's a symlink
}

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

// BuildFileTree generates a string representation of the file tree.
func BuildFileTree(rootDir string, customIgnorePatterns []string, respectGitIgnore bool) (string, error) {
	files, err := GetProjectFiles(rootDir, customIgnorePatterns, respectGitIgnore, nil, true)
	if err != nil {
		return "", fmt.Errorf("failed to get project files for tree: %w", err)
	}

	if len(files) == 0 {
		return "Project is empty or all files are ignored.", nil
	}

	absRootDir, _ := filepath.Abs(rootDir)

	var treeBuilder strings.Builder
	treeBuilder.WriteString(filepath.Base(absRootDir) + "/\n")

	dirContents := make(map[string][]string)
	pathInfos := make(map[string]FileInfo)

	for _, fi := range files {
		pathInfos[fi.RelPath] = fi
		parentDir := filepath.Dir(fi.RelPath)
		if parentDir == "." {
			parentDir = ""
		}
		dirContents[parentDir] = append(dirContents[parentDir], fi.RelPath)
	}

	for k := range dirContents {
		sort.Strings(dirContents[k])
	}

	var buildSubTree func(dirRelPath string, indent string)
	buildSubTree = func(currentDirRelPath string, indent string) {
		key := currentDirRelPath
		if currentDirRelPath == "." || currentDirRelPath == absRootDir || currentDirRelPath == string(filepath.Separator) {
			key = ""
		}

		items := dirContents[key]
		for i, itemRelPath := range items {
			fi, ok := pathInfos[itemRelPath]
			if !ok {
				fmt.Fprintf(os.Stderr, "Warning: FileInfo not found for path %s in tree building\n", itemRelPath)
				continue
			}
			baseName := filepath.Base(itemRelPath)

			prefix := indent
			if i == len(items)-1 {
				prefix += "└── "
			} else {
				prefix += "├── "
			}
			treeBuilder.WriteString(prefix)
			treeBuilder.WriteString(baseName)
			if fi.IsDir {
				treeBuilder.WriteString("/")
			}
			if fi.IsSymlink {
				treeBuilder.WriteString(" (symlink)")
			}
			treeBuilder.WriteString("\n")

			if fi.IsDir {
				newIndent := indent
				if i == len(items)-1 {
					newIndent += "    "
				} else {
					newIndent += "│   "
				}
				buildSubTree(itemRelPath, newIndent)
			}
		}
	}

	buildSubTree("", "")
	return treeBuilder.String(), nil
}

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

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BuildFileTree generates a string representation of the file tree.
func BuildFileTree(rootDir string, customIgnorePatterns []string, respectGitIgnore bool) (string, error) {
	files, err := GetProjectFiles(rootDir, customIgnorePatterns, respectGitIgnore, nil, true) // Calls GetProjectFiles
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
	pathInfos := make(map[string]FileInfo) // Uses FileInfo from types.go

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

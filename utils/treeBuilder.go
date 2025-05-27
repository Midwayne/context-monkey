package utils

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// tree structure (intenarl)
type TreeNode struct {
	Name     string
	IsDir    bool
	Children map[string]*TreeNode
}

// BuildFileTree generates a string representation of the file tree.
func BuildFileTree(
	rootDir string, customIgnorePatterns []string,
	respectGitIgnore bool, specificFileArgs []string, includeDirsInResult bool) (string, error) {

	files, err := GetProjectFiles(rootDir, customIgnorePatterns, respectGitIgnore, specificFileArgs, includeDirsInResult)
	if err != nil {
		return "", fmt.Errorf("failed to get project files: %w", err)
	}

	if len(files) == 0 {
		return "Project is empty or all files are ignored.", nil
	}

	root := &TreeNode{
		Name:     filepath.Base(rootDir),
		IsDir:    true,
		Children: make(map[string]*TreeNode),
	}

	// Build tree
	for _, fi := range files {
		parts := strings.Split(fi.RelPath, string(filepath.Separator))
		curr := root
		for i, part := range parts {
			if part == "." || part == "" {
				continue
			}
			isLast := i == len(parts)-1

			child, exists := curr.Children[part]
			if !exists {
				child = &TreeNode{
					Name:     part,
					IsDir:    fi.IsDir && isLast,
					Children: make(map[string]*TreeNode),
				}
				// if it's an intermediate dir not explicitly returned, mark as dir
				if !isLast {
					child.IsDir = true
				}
				curr.Children[part] = child
			}
			curr = child
		}
	}

	// Render tree
	var b strings.Builder
	b.WriteString(root.Name + "/\n")

	var render func(node *TreeNode, prefix string, isLast bool)
	render = func(node *TreeNode, prefix string, isLast bool) {
		children := make([]*TreeNode, 0, len(node.Children))
		for _, child := range node.Children {
			children = append(children, child)
		}
		sort.Slice(children, func(i, j int) bool {
			return children[i].Name < children[j].Name
		})

		for i, child := range children {
			connector := "├── "
			nextPrefix := prefix + "│   "
			if i == len(children)-1 {
				connector = "└── "
				nextPrefix = prefix + "    "
			}
			b.WriteString(prefix + connector + child.Name)
			if child.IsDir {
				b.WriteString("/")
			}
			b.WriteString("\n")

			if len(child.Children) > 0 {
				render(child, nextPrefix, i == len(children)-1)
			}
		}
	}

	render(root, "", true)
	return b.String(), nil
}

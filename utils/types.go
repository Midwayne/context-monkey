package utils

// FileInfo holds path information for a file.
type FileInfo struct {
	AbsPath   string // Absolute path to the file
	RelPath   string // Path relative to the project root
	IsDir     bool   // True if it's a directory
	IsSymlink bool   // True if it's a symlink
}

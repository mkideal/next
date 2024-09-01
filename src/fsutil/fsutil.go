// Package fsutil provides file system utility functions.
package fsutil

import (
	"os"
	"path/filepath"
)

// AppendFiles appends files with the specified extension to the given slice.
// It searches for files in the specified path and optionally in subdirectories.
//
// Parameters:
//   - files: A slice of strings to which the found file paths will be appended.
//   - path: The directory path to search for files.
//   - ext: The file extension to match (including the dot, e.g., ".txt").
//   - recursive: If true, search subdirectories recursively.
//
// Returns:
//   - A slice of strings containing the original files plus the newly found files.
//   - An error if any occurred during the file system operations.
func AppendFiles(files []string, path, ext string, recursive bool) ([]string, error) {
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(currentPath) == ext {
			files = append(files, currentPath)
		}
		if !recursive && info.IsDir() && currentPath != path {
			return filepath.SkipDir
		}
		return nil
	})
	return files, err
}

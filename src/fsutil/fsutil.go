package fsutil

import (
	"os"
	"path/filepath"
)

func AppendFiles(files []string, path, ext string, recursive bool) ([]string, error) {
	var err = filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
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

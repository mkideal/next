package fsutil

import (
	"os"
	"path/filepath"
)

func AppendFiles(files []string, path, ext string, recursive bool) ([]string, error) {
	var err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ext {
			files = append(files, path)
		}
		if !recursive && info.IsDir() && path != "." && path != ".." {
			return filepath.SkipDir
		}
		return nil
	})
	return files, err
}

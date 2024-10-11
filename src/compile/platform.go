package compile

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// FileSystem is an interface that abstracts the file system functions.
type FileSystem interface {
	fs.FS
	Abs(name string) (string, error)
}

// Platform is an interface that abstracts the platform-specific functions.
type Platform interface {
	// Getenv retrieves the value of the environment variable named by the key.
	Getenv(string) string

	// UserHomeDir returns the current user's home directory.
	UserHomeDir() (string, error)

	// Stdin returns the standard input.
	Stdin() io.Reader

	// Stderr returns the standard error.
	Stderr() io.Writer

	// ReadFile reads the file named by filename and returns the contents.
	ReadFile(string) ([]byte, error)

	// WriteFile writes data to the file named by filename.
	WriteFile(string, []byte, os.FileMode) error

	// IsExist reports whether the named file or directory exists.
	IsExist(string) bool

	// IsNotExist reports whether the named file or directory does not exist.
	IsNotExist(string) bool
}

var _ Platform = standardPlatform{}

// StandardPlatform returns a Platform that uses the standard library functions.
func StandardPlatform() Platform {
	return standardPlatform{}
}

type standardPlatform struct{}

func (standardPlatform) Getenv(key string) string {
	return os.Getenv(key)
}

func (standardPlatform) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (standardPlatform) Stdin() io.Reader {
	return os.Stdin
}

func (standardPlatform) Stderr() io.Writer {
	return os.Stderr
}

func (standardPlatform) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (p standardPlatform) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// Ensure the file is writable.
	info, err := os.Stat(filename)
	if err == nil && info.Mode()&0600 != 0600 {
		if err := os.Chmod(filename, perm|0600); err != nil {
			return err
		}
	}
	return os.WriteFile(filename, data, perm)
}

func (p standardPlatform) IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func (standardPlatform) IsNotExist(filename string) bool {
	_, err := os.Stat(filename)
	return os.IsNotExist(err)
}

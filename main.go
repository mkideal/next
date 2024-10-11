package main

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mkideal/next/src/compile"
)

//go:embed builtin/*
var builtin embed.FS

type embedFS struct {
	embed.FS
}

func (e embedFS) Abs(name string) (string, error) {
	return name, nil
}

type dirFS struct {
	baseDir string
	fs.FS
}

func (d dirFS) Abs(name string) (string, error) {
	return filepath.Join(d.baseDir, name), nil
}

func copyBuiltin() compile.FileSystem {
	builtin := embedFS{FS: builtin}
	if x := strings.ToLower(os.Getenv(compile.NEXTNOCOPYBUILTIN)); x == "1" || x == "true" || x == "on" || x == "yes" {
		return builtin
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return builtin
	}
	builtinBaseDir := filepath.Join(homeDir, ".next")
	builtinDir := filepath.Join(builtinBaseDir, "builtin")
	if err := os.MkdirAll(builtinDir, 0755); err != nil {
		return builtin
	}
	entries, err := builtin.ReadDir("builtin")
	if err != nil {
		return builtin
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := builtin.ReadFile("builtin/" + entry.Name())
		if err != nil {
			return builtin
		}
		path := filepath.Join(builtinDir, entry.Name())
		if content, err := os.ReadFile(path); err == nil && string(content) == string(data) {
			continue
		}
		os.Chmod(path, 0644)
		if err := os.WriteFile(path, data, 0644); err != nil {
			return builtin
		}
		os.Chmod(path, 0444)
	}
	return dirFS{
		baseDir: builtinBaseDir,
		FS:      os.DirFS(builtinBaseDir),
	}
}

func main() {
	compile.Compile(compile.StandardPlatform(), copyBuiltin(), os.Args)
}

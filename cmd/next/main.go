package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gopherd/next/parser"
	"github.com/gopherd/next/types"
)

var flags struct {
	version bool
	debug   bool
}

func main() {
	flag.BoolVar(&flags.version, "version", false, "show version")
	flag.BoolVar(&flags.debug, "debug", false, "enable debug mode")
	flag.Parse()

	var err error
	var files = flag.Args()
	if len(files) == 0 {
		files, err = listAllFileInDir(".")
	} else {
		if len(files) == 1 && filepath.Ext(files[0]) != ".next" {
			files, err = listAllFileInDir(files[0])
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no source files")
		os.Exit(1)
	}

	ctx := types.NewContext(flags.debug)
	for _, file := range files {
		f, err := parser.ParseFile(ctx.FileSet(), file, nil, parser.ParseComments|parser.AllErrors)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := ctx.AddFile(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := ctx.Resolve(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func listAllFileInDir(dir string) ([]string, error) {
	var files []string
	var err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".next" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

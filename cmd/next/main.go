package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopherd/next/parser"
	"github.com/gopherd/next/scanner"
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
		files, err = listDir(".")
	} else {
		if len(files) == 1 && filepath.Ext(files[0]) != ".next" {
			files, err = listDir(files[0])
		}
	}
	if err != nil {
		exit(err)
	}
	if len(files) == 0 {
		exit("no source files")
	}

	ctx := types.NewContext(flags.debug)
	for _, file := range files {
		f, err := parser.ParseFile(ctx.FileSet(), file, nil, parser.ParseComments|parser.AllErrors)
		if err != nil {
			exit(err)
		}
		if err := ctx.AddFile(f); err != nil {
			exit(err)
		}
	}
	if err := ctx.Resolve(); err != nil {
		exit(err)
	}
}

func exit(err any) {
	switch err := err.(type) {
	case nil:
		os.Exit(0)
	case error:
		if err, ok := err.(scanner.ErrorList); ok {
			const maxErrorCount = 20
			var sb strings.Builder
			for i := 0; i < len(err) && i < maxErrorCount; i++ {
				fmt.Fprintln(&sb, err[i])
			}
			if remaining := len(err) - maxErrorCount; remaining > 0 {
				fmt.Fprintf(&sb, "and %d more errors\n", remaining)
			}
			fmt.Fprint(os.Stderr, sb.String())
		}
		os.Exit(2)
	case string:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	case int:
		os.Exit(err)
	default:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func listDir(dir string) ([]string, error) {
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

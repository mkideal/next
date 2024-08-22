package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopherd/core/builder"
	"github.com/gopherd/next/parser"
	"github.com/gopherd/next/scanner"
	"github.com/gopherd/next/types"
)

var flags struct {
	version bool
	debug   bool
}

func main() {
	flag.BoolVar(&flags.version, "version", false, "Show next compiler version")
	flag.BoolVar(&flags.debug, "d", false, "Enable debug mode")
	flag.Parse()

	if flags.version {
		builder.PrintInfo()
	}

	var err error
	var files []string
	if flag.NArg() == 0 {
		files, err = appendFiles(files, ".")
	} else {
		for _, arg := range flag.Args() {
			files, err = appendFiles(files, arg)
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		exit(err)
	}
	if len(files) == 0 {
		exit("no source files")
	}
	seen := make(map[string]bool)
	for i := len(files) - 1; i >= 0; i-- {
		files[i], err = filepath.Abs(files[i])
		if err != nil {
			exit(err)
		}
		if seen[files[i]] {
			// remove duplicated files
			files = append(files[:i], files[i+1:]...)
		} else {
			seen[files[i]] = true
		}
	}
	debugf("files: %v\n", files)

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

func debugf(format string, args ...interface{}) {
	if flags.debug {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func exit(err any) {
	switch err := err.(type) {
	case nil:
		os.Exit(0)
	case error:
		if errs, ok := err.(scanner.ErrorList); ok {
			const maxErrorCount = 20
			var sb strings.Builder
			for i := 0; i < len(errs) && i < maxErrorCount; i++ {
				fmt.Fprintln(&sb, errs[i])
			}
			if remaining := len(errs) - maxErrorCount; remaining > 0 {
				fmt.Fprintf(&sb, "and %d more errors\n", remaining)
			}
			fmt.Fprint(os.Stderr, sb.String())
		} else {
			fmt.Fprintln(os.Stderr, err)
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

func appendFiles(files []string, path string) ([]string, error) {
	var err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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

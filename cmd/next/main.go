package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopherd/core/builder"
	"github.com/gopherd/next/internal/fsutil"
	"github.com/gopherd/next/parser"
	"github.com/gopherd/next/scanner"
	"github.com/gopherd/next/types"
)

func main() {
	ctx := types.NewContext()
	version := flag.Bool("v", false, "Show next compiler version")
	ctx.SetupCommandFlags(flag.CommandLine)
	flag.Parse()

	if *version {
		builder.PrintInfo()
		os.Exit(0)
	}

	var files []string
	if flag.NArg() == 0 {
		files = result(fsutil.AppendFiles(files, ".", ".next", false))
	} else {
		for _, arg := range flag.Args() {
			files = result(fsutil.AppendFiles(files, arg, ".next", false))
		}
	}
	if len(files) == 0 {
		try("no source files")
	}

	// compute absolute path and remove duplicated files
	seen := make(map[string]bool)
	for i := len(files) - 1; i >= 0; i-- {
		files[i] = result(filepath.Abs(files[i]))
		if seen[files[i]] {
			files = append(files[:i], files[i+1:]...)
		} else {
			seen[files[i]] = true
		}
	}

	// parse and resolve all files
	for _, file := range files {
		f := result(parser.ParseFile(ctx.FileSet(), file, nil, parser.ParseComments|parser.AllErrors))
		try(ctx.AddFile(f))
	}
	try(ctx.Resolve())

	// generate files
	try(ctx.Generate())
}

func result[T any](v T, err error) T {
	try(err)
	return v
}

func try(err any) {
	switch err := err.(type) {
	case nil:
		return // do nothing if no error
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

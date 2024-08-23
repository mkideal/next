package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopherd/core/builder"
	"github.com/gopherd/core/flags"
	"github.com/gopherd/next/internal/fsutil"
	"github.com/gopherd/next/parser"
	"github.com/gopherd/next/scanner"
	"github.com/gopherd/next/types"
)

const currentDir = "."
const nextExt = ".next"

func main() {
	ctx := types.NewContext()
	ctx.SetupCommandFlags(flag.CommandLine, flags.UseUsage(flag.CommandLine.Output()))

	flag.Usage = func() {
		var w strings.Builder
		fmt.Fprintf(&w, "Usage: %s [Options] [dirs or files...]\n", os.Args[0])
		fmt.Fprintf(&w, "       %s [Options] [stdin]\n", os.Args[0])
		fmt.Fprintf(&w, "       %s version\n", os.Args[0])
		fmt.Fprintf(&w, "\nOptions:\n")
		fmt.Fprintf(flag.CommandLine.Output(), w.String())
		flag.CommandLine.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 1 && flag.Arg(0) == "version" {
		builder.PrintInfo()
		os.Exit(0)
	}

	var files []string
	var stdin io.Reader
	if flag.NArg() == 0 {
		if stat, err := os.Stdin.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			stdin = os.Stdin
			files = append(files, "<stdin>")
		} else {
			files = result(fsutil.AppendFiles(files, currentDir, nextExt, false))
		}
	} else {
		for _, arg := range flag.Args() {
			if arg == "-" {
				try("invalid argument: -")
			}
			if strings.HasPrefix(arg, "-") {
				try(fmt.Sprintf("flag %q not allowed after non-flag argument", arg))
			}
		}
		for _, arg := range flag.Args() {
			files = result(fsutil.AppendFiles(files, arg, nextExt, false))
		}
	}
	if len(files) == 0 {
		try("no source files")
	}

	// compute absolute path and remove duplicated files
	if stdin == nil {
		seen := make(map[string]bool)
		for i := len(files) - 1; i >= 0; i-- {
			files[i] = result(filepath.Abs(files[i]))
			if seen[files[i]] {
				files = append(files[:i], files[i+1:]...)
			} else {
				seen[files[i]] = true
			}
		}
	}

	// parse and resolve all files
	for _, file := range files {
		f := result(parser.ParseFile(ctx.FileSet(), file, stdin, parser.ParseComments|parser.AllErrors))
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

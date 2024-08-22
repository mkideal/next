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

type Map map[string]string

func (m *Map) Set(s string) error {
	if *m == nil {
		*m = make(Map)
	}
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: %q, expected key=value", s)
	}
	(*m)[parts[0]] = parts[1]
	return nil
}

func (m Map) String() string {
	if m == nil {
		return ""
	}
	var sb strings.Builder
	for k, v := range m {
		if sb.Len() > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, "%s=%s", k, v)
	}
	return sb.String()
}

func main() {
	ctx := types.NewContext()
	version := flag.Bool("v", false, "Show next compiler version")
	ctx.SetupCommandFlags(flag.CommandLine)
	flag.Parse()

	if *version {
		builder.PrintInfo()
		os.Exit(0)
	}

	var err error
	var files []string
	if flag.NArg() == 0 {
		files, err = fsutil.AppendFiles(files, ".", ".next", false)
	} else {
		for _, arg := range flag.Args() {
			files, err = fsutil.AppendFiles(files, arg, ".next", false)
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

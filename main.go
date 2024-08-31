package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopherd/core/builder"
	"github.com/gopherd/core/flags"
	"github.com/gopherd/core/term"

	"github.com/next/next/src/fsutil"
	"github.com/next/next/src/parser"
	"github.com/next/next/src/scanner"
	"github.com/next/next/src/types"
)

//go:embed builtin/*
var builtin embed.FS

const currentDir = "."
const nextExt = ".next"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		builder.PrintInfo()
		os.Exit(0)
	}

	flagSet := flag.CommandLine
	flagSet.Init(os.Args[0], flag.ContinueOnError)
	flagSet.Usage = func() {}

	ctx := types.NewContext(builtin)
	ctx.SetupCommandFlags(flagSet, flags.UseUsage(flagSet.Output(), flags.NameColor(term.Bold)))

	// set output color for error messages
	flagSet.SetOutput(term.ColorizeWriter(os.Stderr, term.Red))
	usage := func() {
		var w strings.Builder
		flagSet.SetOutput(os.Stderr)
		fmt.Fprint(&w, "Next is an IDL for generating customized code across multiple languages.\n\n")
		fmt.Fprintf(&w, "Usage: %s [Options] [source dirs and/or files... (default .)]\n", os.Args[0])
		fmt.Fprintf(&w, "       %s [Options] <stdin>\n", os.Args[0])
		fmt.Fprintf(&w, "       %s version\n", os.Args[0])
		fmt.Fprintf(&w, "\nOptions:\n")
		fmt.Fprint(flagSet.Output(), w.String())
		flagSet.PrintDefaults()
		term.Fprintf(flagSet.Output(), `
For more information:
  Website:    %s
  Repository: %s

`,
			(term.Bold + term.BrightBlue).Colorize("https://nextlang.org"),
			(term.Bold + term.BrightBlue).Colorize("https://github.com/next/next"),
		)
	}
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			usage()
			os.Exit(0)
		}
		usageError(flagSet, "")
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
				usageError(flagSet, "invalid argument: -")
			}
			if strings.HasPrefix(arg, "-") {
				usageError(flagSet, "flag %q not allowed after non-flag argument", arg)
			}
		}
		for _, arg := range flag.Args() {
			files = result(fsutil.AppendFiles(files, arg, nextExt, false))
		}
	}
	if len(files) == 0 {
		usageError(flagSet, "no source files")
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

func usageError(flagSet *flag.FlagSet, format string, args ...any) {
	if format != "" {
		fmt.Fprintln(flagSet.Output(), fmt.Sprintf(format, args...))
	}
	try(fmt.Errorf("try %q for help", os.Args[0]+" -h"))
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

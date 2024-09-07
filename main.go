package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

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
const website = "https://nextlang.org"
const repository = "https://github.com/next/next"

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
		flagSet.SetOutput(os.Stderr)
		name := term.Bold.Colorize(os.Args[0])
		term.Fprint(flagSet.Output(), "Next is an IDL for generating customized code across multiple languages.\n\n")
		term.Fprint(flagSet.Output(), "Usage:\n")
		term.Fprintf(flagSet.Output(), "  %s [Options] [source_dirs_or_files...] (default: current directory)\n", name)
		term.Fprintf(flagSet.Output(), "  %s [Options] <stdin>\n", name)
		term.Fprintf(flagSet.Output(), "  %s version\n", name)
		term.Fprintf(flagSet.Output(), "\nOptions:\n")
		flagSet.PrintDefaults()
		term.Fprintf(flagSet.Output(), `For more information:
  Website:    %s
  Repository: %s

`,
			(term.Bold + term.BrightBlue).Colorize(website),
			(term.Bold + term.BrightBlue).Colorize(repository),
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
		f := result(parser.ParseFile(ctx.FileSet(), file, stdin, parser.ParseComments))
		result(ctx.AddFile(f))
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
			fmt.Fprintln(os.Stderr, tryExtractTemplateTrace(err))
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

func tryExtractTemplateTrace(err error) error {
	if err == nil {
		return nil
	}
	var errs []error
	for err != nil {
		if e, ok := err.(template.ExecError); ok {
			errs = append(errs, err)
			err = e.Err
		} else if e := errors.Unwrap(err); e != nil {
			err = e
		} else {
			errs = append(errs, err)
			break
		}
	}
	for i := 0; i+1 < len(errs); i++ {
		indent := strings.Repeat(" ", len(errs)-i-1)
		errs[i] = errors.New(indent + strings.TrimSuffix(strings.TrimPrefix(strings.TrimSuffix(
			strings.TrimSpace(errs[i].Error()),
			strings.TrimSpace(errs[i+1].Error()),
		), "template: "), ": "))
	}
	slices.Reverse(errs)
	return errors.Join(errs...)
}

//go:generate go run github.com/gopherd/tools/cmd/docgen@v0.0.8 -ext .mdx -I ./ -o ../../website/docs/api/preview -level 0 -M "---" -M "pagination_prev: null" -M "pagination_next: null" -M "---"
package compile

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/gopherd/core/builder"
	"github.com/gopherd/core/encoding"
	"github.com/gopherd/core/flags"
	"github.com/gopherd/core/op"
	"github.com/gopherd/core/term"
	"gopkg.in/yaml.v3"

	"github.com/next/next/src/grammar"
	"github.com/next/next/src/internal/fsutil"
	"github.com/next/next/src/parser"
	"github.com/next/next/src/scanner"
)

// @api(Environment/NEXTNOCOPYBUILTIN) represents the environment variable to disable copying builtin files.
// If the value is `1`, `true`, `on`, or `yes` (case-insensitive), builtin files will not be copied to the user home directory.
const NEXTNOCOPYBUILTIN = "NEXTNOCOPYBUILTIN"

// @api(Environment/NEXTMAXSTACK) represents the environment variable to set the maximum stack depth.
// The value is a positive integer that represents the maximum stack depth. The default value is 100.
const NEXTMAXSTACK = "NEXTMAXSTACK"

const nextExt = ".next"
const website = "https://next.as"
const repository = "https://github.com/next/next"

type commandContext struct {
	platform Platform
	builtin  FileSystem
}

type command struct {
	description string
	run         func(*commandContext, []string) error
}

func newCommand(desc string, run func(*commandContext, []string) error) *command {
	return &command{description: desc, run: run}
}

var commands = map[string]*command{
	// @api(CommandLine/Command/version) command prints the version of the next compiler.
	// It prints the version of the next compiler and exits the program.
	//
	// Example:
	//
	//	```sh
	//	next version
	//	```
	//
	// Output:
	//
	//	```
	//	next v0.0.4(main: 51864a35de7890d63bfd8acecdb62d20372ca963) built at 2024/09/27T22:58:21+0800 by go1.23.0
	//	```
	"version": newCommand(
		"Print the version of the next compiler",
		func(ctx *commandContext, _ []string) error {
			builder.PrintInfo()
			unwrap(ctx.platform.Stderr(), 0)
			return nil
		},
	),

	// @api(CommandLine/Command/grammar) command generates the default grammar for the next files.
	//
	// Example:
	//
	//	```sh
	//	next grammar # generate the default grammar to the standard output with YAML format
	//	next grammar grammar.yaml # generate the default grammar with YAML format
	//	next grammar grammar.yml # generate the default grammar with YAML format
	//	next grammar grammar.json # generate the default grammar with JSON format
	//	```
	//
	//	:::note
	//	The default grammar is generated in YAML format if the file extension is not provided.
	//	Currently, the supported file extensions are `.json`, `.yaml`, and `.yml` (alias of `.yaml`).
	//	:::
	"grammar": newCommand(
		"Generate the default grammar for the next files: next grammar [filename]",
		func(_ *commandContext, args []string) error {
			var path string
			var ext string
			if len(args) > 2 {
				return fmt.Errorf("too many arguments: next %s [filename]", args[0])
			}
			if len(args) == 2 {
				path = args[1]
				ext = filepath.Ext(path)
			}

			content, err := json.Marshal(grammar.Builtin)
			if err == nil {
				switch ext {
				case ".json":
					content, err = encoding.Transform(content, json.Unmarshal, func(x any) ([]byte, error) {
						return json.MarshalIndent(x, "", "  ")
					})
				case "", ".yaml", ".yml":
					var buf bytes.Buffer
					enc := yaml.NewEncoder(&buf)
					enc.SetIndent(2)
					content, err = encoding.Transform(content, json.Unmarshal, func(x any) ([]byte, error) {
						if err := enc.Encode(x); err != nil {
							return nil, err
						}
						return buf.Bytes(), nil
					})
				default:
					return fmt.Errorf("unsupported file extension: %q, use .json, .yaml, or .yml", ext)
				}
			}
			if err != nil {
				return err
			}
			if path != "" {
				return os.WriteFile(path, content, 0644)
			}
			fmt.Println(string(content))
			return nil
		},
	),

	// @api(CommandLine/Command/build) command builds a next project file or runs all next project files in a directory.
	// It reads the project file and compiles the source files according to the project configuration.
	// The project file is a YAML or JSON file that contains the project [Configuration](#CommandLine/Configuration).
	//
	// Example:
	//
	//	```sh
	//	# build all .nextproj files in the current directory
	//	next build
	//
	//	# build the example.nextproj file
	//	next build example.nextproj
	//
	//	# build all .nextproj files in the example directory
	//	next build example/
	//
	//	# build multiple project files or directories
	//	next build example1.nextproj example2.nextproj example
	//	```
	//
	//	:::tip
	//	**.nextproj** is recommended for the Next project file.
	//	:::
	"build": newCommand(
		"Run a next project: next build [file_or_dirs...]",
		func(ctx *commandContext, args []string) error {
			if len(args) == 1 {
				args = append(args, ".")
			}
			var projects []string
			for _, arg := range args[1:] {
				files, err := fsutil.AppendFiles(projects, arg, ".nextproj", true)
				if err != nil {
					return fmt.Errorf("failed to read %q: %v", arg, err)
				}
				projects = append(projects, files...)
			}
			if len(projects) == 0 {
				return fmt.Errorf("no project files")
			}
			for _, project := range projects {
				content, err := os.ReadFile(project)
				if err != nil {
					return fmt.Errorf("failed to read %q: %w", project, err)
				}
				compiler := NewCompiler(ctx.platform, ctx.builtin)
				var options struct {
					*Options `yaml:",inline"`

					// @api(CommandLine/Configuration.sources) represents the source directories or files.
					//
					// Example:
					//
					//	```yaml
					//	sources:
					//	  - demo.next
					//	  - src/next/
					//	```
					Sources []string `yaml:"sources" json:"sources"`
				}
				options.Options = &compiler.options
				var unmarshaler func([]byte, any) error
				ext := filepath.Ext(project)
				switch ext {
				case ".json":
					unmarshaler = json.Unmarshal
				case "", ".yaml", ".yml", ".nextproj":
					unmarshaler = yaml.Unmarshal
				default:
					return fmt.Errorf("unsupported file extension: %q, use .json, .yaml, .yml, or .nextproj (alias of .yaml)", ext)
				}
				if err := unmarshaler(content, &options); err != nil {
					return fmt.Errorf("failed to parse %q: %v", project, err)
				}
				options.resolvePath(filepath.Dir(project))
				var files []string
				for _, arg := range options.Sources {
					arg = filepath.Join(filepath.Dir(project), arg)
					file, err := fsutil.AppendFiles(files, arg, nextExt, false)
					if err != nil {
						return fmt.Errorf("failed to read %q: %v", arg, err)
					}
					files = result(compiler.platform.Stderr(), file, err)
				}
				if len(files) == 0 {
					term.Fprintf(ctx.platform.Stderr(), "%s: %s\n", project, term.BrightYellow.Colorize("no source files"))
					continue
				}
				doCompile(compiler, files, nil)
			}
			return nil
		},
	),
}

// Compile compiles the next files.
func Compile(platform Platform, builtin FileSystem, args []string) {
	ctx := &commandContext{platform: platform, builtin: builtin}

	if len(args) >= 2 {
		if cmd, ok := commands[args[1]]; ok {
			unwrap(ctx.platform.Stderr(), cmd.run(ctx, args[1:]))
			unwrap(ctx.platform.Stderr(), 0)
		}
	}

	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flagSet.Usage = func() {}

	stdin, stderr := platform.Stdin(), platform.Stderr()
	compiler := NewCompiler(platform, builtin)
	compiler.SetupCommandFlags(flagSet, flags.UseUsage(flagSet.Output(), flags.NameColor(term.Bold)))

	// set output color for error messages
	flagSet.SetOutput(term.ColorizeWriter(stderr, term.Red))
	usageFunc := func() {
		flagSet.SetOutput(stderr)
		name := term.Bold.Colorize(args[0])
		term.Fprint(flagSet.Output(), "Next is an IDL for generating customized code across multiple languages.\n\n")
		term.Fprint(flagSet.Output(), "Usage:\n")
		term.Fprintf(flagSet.Output(), "  %s [Options] [source_dirs_or_files...] (default: current directory)\n", name)
		term.Fprintf(flagSet.Output(), "  %s [Options] <stdin>\n", name)
		term.Fprintf(flagSet.Output(), "  %s <Command> [Command-Args]\n", name)
		term.Fprintf(flagSet.Output(), "\nCommands:\n")

		var maxCommandLength int
		for _, name := range slices.Sorted(maps.Keys(commands)) {
			maxCommandLength = max(maxCommandLength, len(name))
		}
		maxCommandLength += 4 // add 4 spaces for padding
		for _, name := range slices.Sorted(maps.Keys(commands)) {
			term.Fprintf(flagSet.Output(), "  %s%s%s\n", term.BrightMagenta.Colorize(name), strings.Repeat(" ", maxCommandLength-len(name)), commands[name].description)
		}

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
	if err := flagSet.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			usageFunc()
			unwrap(stderr, 0)
		}
		usage(flagSet, stderr, "")
	}

	var files []string
	var source io.Reader
	if flagSet.NArg() == 0 {
		if osStdin, ok := stdin.(*os.File); ok {
			if stat, err := osStdin.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
				source = osStdin
				files = append(files, "<stdin>")
			} else {
				newFiles, err := fsutil.AppendFiles(files, ".", nextExt, false)
				files = result(stderr, newFiles, err)
			}
		} else if stdin != nil {
			source = stdin
			files = append(files, "<input>")
		}
	} else {
		for _, arg := range flagSet.Args() {
			if arg == "-" {
				usage(flagSet, stderr, "invalid argument: -")
			}
			if strings.HasPrefix(arg, "-") {
				usage(flagSet, stderr, "flag %q not allowed after non-flag argument", arg)
			}
		}
		for _, arg := range flagSet.Args() {
			path, err := fsutil.AppendFiles(files, arg, nextExt, false)
			files = result(stderr, path, err)
		}
	}

	if len(files) == 0 {
		usage(flagSet, stderr, "no source files")
	}

	doCompile(compiler, files, source)
}

func doCompile(compiler *Compiler, files []string, source io.Reader) {
	platform := compiler.platform
	stderr := compiler.platform.Stderr()

	// compute absolute path and remove duplicated files
	if source == nil {
		seen := make(map[string]bool)
		for i := len(files) - 1; i >= 0; i-- {
			path, err := filepath.Abs(files[i])
			files[i] = result(stderr, path, err)
			if seen[files[i]] {
				files = append(files[:i], files[i+1:]...)
			} else {
				seen[files[i]] = true
			}
		}
	}

	// read grammar file
	if compiler.options.Grammar != "" {
		content, err := platform.ReadFile(compiler.options.Grammar)
		content = result(stderr, content, err)
		switch ext := filepath.Ext(compiler.options.Grammar); ext {
		case ".json":
		case "", ".yaml", ".yml":
			content, err = encoding.Transform(content, yaml.Unmarshal, json.Marshal)
			content = result(stderr, content, err)
		default:
			unwrap(stderr, fmt.Sprintf("unsupported file extension: %q, use .json, .yaml, or .yml", ext))
		}
		unwrap(stderr, json.Unmarshal(content, &compiler.grammar))
		unwrap(stderr, compiler.grammar.Resolve())
	} else {
		compiler.grammar = grammar.Builtin
	}

	// parse and resolve all files
	for _, file := range files {
		if source == nil {
			content, err := platform.ReadFile(file)
			source = bytes.NewReader(result(stderr, content, err))
		}
		f, err := parser.ParseFile(compiler.FileSet(), file, source, parser.ParseComments)
		unwrap(stderr, op.Second(compiler.AddFile(result(stderr, f, err))))
		source = nil
	}
	unwrap(stderr, compiler.Resolve())

	// Validate the grammar
	unwrap(stderr, compiler.ValidateGrammar())

	// generate files
	unwrap(stderr, Generate(compiler))
}

// exit exits the program. It is a variable for overriding.
var exit = os.Exit

// usage prints usage message and exits the program. It is a variable for overriding.
var usage = func(flagSet *flag.FlagSet, stderr io.Writer, format string, args ...any) {
	if format != "" {
		fmt.Fprintln(flagSet.Output(), fmt.Sprintf(format, args...))
	}
	unwrap(stderr, fmt.Errorf("try %q for help", flagSet.Name()+" -h"))
}

// result returns the value if err is nil, otherwise it unwraps the error and exits the program.
func result[T any](stderr io.Writer, v T, err error) T {
	unwrap(stderr, err)
	return v
}

// unwrap unwraps error and exits the program if error is not nil.
func unwrap(stderr io.Writer, err any) {
	switch err := err.(type) {
	case nil:
		return // do nothing if no error
	case error:
		if errs, ok := err.(scanner.ErrorList); ok {
			const maxErrorCount = 20
			for i := 0; i < len(errs) && i < maxErrorCount; i++ {
				printErrorWithPosition(stderr, errs[i].Error())
			}
			if remaining := len(errs) - maxErrorCount; remaining > 0 {
				fmt.Fprintf(stderr, "and %d more errors\n", remaining)
			}
		} else {
			printError(stderr, err)
		}
		exit(2)
	case string:
		fmt.Fprintln(stderr, err)
		exit(1)
	case int:
		exit(err)
	default:
		fmt.Fprintln(stderr, err)
		exit(1)
	}
}

func printError(stderr io.Writer, err error) {
	if err == nil {
		return
	}
	var errs []string
	origin := err
	for err != nil {
		if e, ok := err.(template.ExecError); ok {
			errs = append(errs, err.Error())
			err = e.Err
			origin = nil
		} else if e := errors.Unwrap(err); e != nil {
			err = e
			if origin == nil {
				origin = err
			}
		} else {
			if origin == nil {
				errs = append(errs, err.Error())
			} else {
				errs = append(errs, origin.Error())
			}
			break
		}
	}
	for i := 0; i+1 < len(errs); i++ {
		errs[i] = strings.TrimSuffix(strings.TrimSpace(errs[i]), strings.TrimSpace(errs[i+1]))
	}
	slices.Reverse(errs)
	for i := len(errs) - 1; i >= 0; i-- {
		s := strings.TrimSpace(errs[i])
		s = strings.TrimPrefix(s, "template: ")
		s = strings.TrimSuffix(s, ":")
		if s == "" || strings.HasPrefix(parseFilename(s), StubPrefix) {
			errs = append(errs[:i], errs[i+1:]...)
		} else {
			errs[i] = s
		}
	}
	if len(errs) == 0 {
		printErrorWithPosition(stderr, err.Error())
		return
	}
	printErrorWithPosition(stderr, errs[0])
	maxIndent := 32
	if len(errs) > maxIndent {
		maxIndent = 0
	}
	for i := 1; i < len(errs); i++ {
		fmt.Fprint(stderr, strings.Repeat(" ", min(i, maxIndent)))
		printErrorWithPosition(stderr, errs[i])
	}
}

func parseFilename(err string) string {
	parts := strings.SplitN(err, ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return parts[0]
}

// printErrorWithPosition tries to print template error in a more readable format.
// template error format: "<filename>:<line>:<column>: <error message>"
func printErrorWithPosition(stderr io.Writer, err string) {
	const fileColor = term.Color("")
	const lineColor = term.BrightBlue
	const columnColor = term.BrightGreen
	const errorColor = term.BrightRed

	if err == "" {
		return
	}
	parts := strings.SplitN(err, ":", 4)
	if len(parts) < 3 {
		term.Fprintln(stderr, errorColor.Colorize(err))
		return
	}
	filename := parts[0]
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, filename); err == nil && !strings.HasPrefix(rel, "..") {
			filename = rel
		}
	}
	line := parts[1]
	column := ""
	if len(parts) > 3 {
		part := parts[2]
		if i, err := strconv.Atoi(part); err == nil {
			if !strings.HasSuffix(filename, nextExt) {
				// add 1 to column if it is a number for non-next files (most likely for template files)
				column = strconv.Itoa(i + 1)
			} else {
				column = part
			}
		}
	}
	message := parts[len(parts)-1]
	if column == "" {
		term.Fprintf(
			stderr, "%s:%s:%s\n",
			fileColor.Colorize(filename),
			lineColor.Colorize(line),
			errorColor.Colorize(message),
		)
	} else {
		term.Fprintf(
			stderr, "%s:%s:%s:%s\n",
			fileColor.Colorize(filename),
			lineColor.Colorize(line),
			columnColor.Colorize(column),
			errorColor.Colorize(message),
		)
	}
}

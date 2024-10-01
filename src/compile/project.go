package compile

import (
	"flag"
	"path/filepath"

	"github.com/gopherd/core/flags"
	"github.com/gopherd/core/term"
)

// @api(CommandLine/Configuration)
//
//	import CodeBlock from "@theme/CodeBlock";
//	import ExampleNextProjSource from "!!raw-loader!@site/example/example.nextproj";
//
// Configuration represents the configuration of the Next project.
// The configuration is used to mamange the compiler options, such as verbosity, output directories, and custom templates.
// If the configuration is provided, you can generate code like this:
//
//	```sh
//	next build example.nextproj
//	```
//
// The configuration file is a YAML or JSON (for .json extension) file that contains the compiler options.
// Here is an example of the configuration file:
//
//	<CodeBlock language="yaml" title="example.nextproj">
//		{ExampleNextProjSource}
//	</CodeBlock>
type Configuration struct {
	test bool `yaml:"-" json:"-"`

	// @api(CommandLine/Configuration.verbose) represents the verbosity level of the compiler.
	//
	// Example:
	//
	//	```yaml
	//	verbose: 1
	//	```
	//
	// See the [-v](#CommandLine/Flag/-v) flag for more information.
	Verbose int `yaml:"verbose" json:"verbose"`

	// @api(CommandLine/Configuration.head) represents the header comment for generated code.
	//
	// Example:
	//
	//	```yaml
	//	head: "Code generated by Next; DO NOT EDIT."
	//	```
	//
	// See the [-head](#CommandLine/Flag/-head) flag for more information.
	Head string `yaml:"head" json:"head"`

	// @api(CommandLine/Configuration.grammar) represents the custom grammar for the next source code.
	//
	// Example:
	//
	//	```yaml
	//	grammar: grammar.yaml
	//	```
	//
	// See the [-g](#CommandLine/Flag/-g) flag for more information.
	Grammar string `yaml:"grammar" json:"grammar"`

	// @api(CommandLine/Configuration.strict) represents the strict mode of the compiler.
	//
	// Example:
	//
	//	```yaml
	//	strict: true
	//	```
	//
	// See the [-s](#CommandLine/Flag/-s) flag for more information.
	Strict bool `yaml:"strict" json:"strict"`

	// @api(CommandLine/Configuration.env) represents the custom environment variables for code generation.
	//
	// Example:
	//
	//	```yaml
	//	env:
	//	  VERSION: "2.1"
	//	  DEBUG: ""
	//	  NAME: myapp
	//	```
	//
	// See the [-D](#CommandLine/Flag/-D) flag for more information.
	Env flags.Map `yaml:"env" json:"env"`

	// @api(CommandLine/Configuration.output) represents the output directories for generated code of each target language.
	//
	// Example:
	//
	//	```yaml
	//	output:
	//	  go: ./output/go
	//	  ts: ./output/ts
	//	```
	//
	// See the [-O](#CommandLine/Flag/-O) flag for more information.
	Output flags.Map `yaml:"output" json:"output"`

	// @api(CommandLine/Configuration.mapping) represents the language-specific type mappings and features.
	//
	// Example:
	//
	//	```yaml
	//	mapping:
	//	  cpp.vector: "std::vector<%T%>"
	//	  java.array: "ArrayList<%T%>"
	//	  go.map: "map[%K%]%V%"
	//	  python.ext: ".py"
	//	  protobuf.vector: "repeated %T.E%"
	//	  ruby.comment: "# %T%"
	//	```
	//
	// See the [-M](#CommandLine/Flag/-M) flag for more information.
	Mapping flags.Map `yaml:"mapping" json:"mapping"`

	// @api(CommandLine/Configuration.templates) represents the custom template directories or files for each target language.
	//
	// Example:
	//
	//	```yaml
	//	templates:
	//	  go:
	//	    - ./templates/go
	//	    - ./templates/go_extra.npl
	//	  python:
	//	    - ./templates/python.npl
	//	```
	//
	// See the [-T](#CommandLine/Flag/-T) flag for more information.
	Templates flags.MapSlice `yaml:"templates" json:"templates"`

	// @api(CommandLine/Configuration.solvers) represents the custom annotation solver programs for code generation.
	//
	// Example:
	//
	//	```yaml
	//	solvers:
	//	  message: "message-type-allocator message-types.json"
	//	```
	//
	// See the [-X](#CommandLine/Flag/-X) flag for more information.
	Solvers flags.MapSlice `yaml:"solvers" json:"solvers"`
}

func (o *Configuration) SetupCommandFlags(flagSet *flag.FlagSet, u flags.UsageFunc) {
	isAnsiSupported := term.IsTerminal(flagSet.Output()) && term.IsSupportsAnsi()
	grey := func(s string) string {
		if isAnsiSupported {
			return term.Gray.Format(s)
		}
		return s
	}
	b := func(s string) string {
		if isAnsiSupported {
			return term.Bold.Format(s)
		}
		return s
	}

	// @api(CommandLine/Flag/-v) represents the verbosity level of the compiler.
	// The default value is **0**, which only shows error messages.
	// The value **1** shows debug messages, and **2** shows trace messages.
	// Usually, the trace message is used for debugging the compiler itself.
	// Levels **1** (debug) and above enable execution of:
	//
	// - **print** and **printf** in Next source files (.next).
	// - **debug** in Next template files (.npl).
	//
	// Example:
	//
	//	```sh
	//	next -v 1 ...
	//	```
	flagSet.IntVar(&o.Verbose, "v", 0, u(""+
		"Control verbosity of compiler output and debugging information.\n"+
		"`VERBOSE` levels: "+b("0")+"=error,  "+b("1")+"=debug, "+b("2")+"=trace.\n"+
		"Usually, the trace message used for debugging the compiler itself.\n"+
		"Levels "+b("1")+" (debug) and above enable execution of:\n"+
		" -"+b("print")+" and "+b("printf")+" in Next source files (.next).\n"+
		" -"+b("debug")+" in Next template files (.npl).\n",
	))

	// @api(CommandLine/Flag/-t) represents the test mode of the compiler.
	// The default value is **false**, which means the compiler is not in test mode.
	// The value **true** enables the test mode, which is used for validating but not generating code.
	//
	// Example:
	//
	//	```sh
	//	next -t ...
	//	```
	flagSet.BoolVar(&o.test, "t", false, u(""+
		"Enable test mode for validating but not generating code.\n"+
		"Example:\n"+
		"  next -t ...\n",
	))

	// @api(CommandLine/Flag/-head) represents the header comment for generated code.
	// The value is a string that represents the header comment for generated code.
	// The default value is an empty string, which means default header comment is used.
	//
	// Example:
	//
	//	```sh
	//	next -head "Code generated by Next; DO NOT EDIT." ...
	//	```
	//
	// :::note
	//
	// Do not add the comment characters like `//` or `/* */` in the header. Next will add them automatically
	// based on the target language comment style.
	//
	// :::
	flagSet.StringVar(&o.Head, "head", "", u(""+
		"Set the `HEAD` comment for each generated file. It will be added to the top of each generated file by {{head}}.\n"+
		"Example:\n"+
		"  next -head \"Code generated by Next; DO NOT EDIT.\"\n",
	))

	// @api(CommandLine/Flag/-g) represents the custom grammar for the next source code.
	//
	// Example:
	//
	//	```sh
	//	next -g grammar.yaml ...
	//	```
	//
	//	:::note
	//	By default, the compiler uses the built-in grammar for the next source code.
	//	You can set a custom grammar file to define a subset of the next grammar.
	//	The grammar file is a JSON file that contains rules.
	//
	//	If `-s` is not set, the compiler will ignore unknown annotations and unknown annotation parameters.
	//	:::
	flagSet.StringVar(&o.Grammar, "g", "", u(""+
		"Set the custom `GRAMMAR` file for the next source code.\n"+
		"The grammar is used to define a subset of the next grammar. It can limit the features of the next code according\n"+
		"by your requirements. The grammar file is a JSON file that contains rules.\n",
	))

	// @api(CommandLine/Flag/-s) represents the strict mode of the compiler.
	// The default value is **false**, which means the compiler is not in strict mode.
	// The value **true** enables the strict mode, which is used for strict validation of the next source code,
	// such as unknown annotations and unknown annotation parameters.
	//
	// Example:
	//
	//	```sh
	//	next -s ...
	//	```
	flagSet.BoolVar(&o.Strict, "s", false, u(""+
		"Enable strict mode for strict validation of the next source code.\n"+
		"Example:\n"+
		"  next -s ...\n",
	))

	// @api(CommandLine/Flag/-D) represents the custom environment variables for code generation.
	// The value is a map of environment variable names and their optional values.
	//
	// Example:
	//
	//	```sh
	//	next -D VERSION=2.1 -D DEBUG -D NAME=myapp ...
	//	```
	//
	//	```npl
	//	{{env.NAME}}
	//	{{env.VERSION}}
	//	```
	//
	// Output:
	//
	//	```
	//	myapp
	//	2.1
	//	```
	flagSet.Var(&o.Env, "D", u(""+
		"Define custom environment variables for use in code generation templates.\n"+
		"`NAME"+grey("[=VALUE]")+"` represents the variable name and its optional value.\n"+
		"Example:\n"+
		"  next -D VERSION=2.1 -D DEBUG -D NAME=myapp\n"+
		"And then, use the variables in templates like this: {{env.NAME}}, {{env.VERSION}}\n",
	))

	// @api(CommandLine/Flag/-O) represents the output directories for generated code of each target language.
	//
	// Example:
	//
	//	```sh
	//	next -O go=./output/go -O ts=./output/ts ...
	//	```
	//
	// :::tip
	//
	// The `{{meta.path}}` is relative to the output directory.
	//
	// :::
	flagSet.Var(&o.Output, "O", u(""+
		"Set output directories for generated code, organized by target language.\n"+
		"`LANG=DIR` specifies the target language and its output directory.\n"+
		"Example:\n"+
		"  next -O go=./output/go -O ts=./output/ts\n",
	))

	// @api(CommandLine/Flag/-T) represents the custom template directories or files for each target language.
	// You can specify multiple templates for a single language.
	//
	// Example:
	//
	//	```sh
	//	next -T go=./templates/go \
	//	     -T go=./templates/go_extra.npl \
	//	     -T python=./templates/python.npl \
	//	     ...
	//	```
	flagSet.Var(&o.Templates, "T", u(""+
		"Specify custom template directories or files for each target language.\n"+
		"`LANG=PATH` defines the target language and its template directory or file.\n"+
		"Multiple templates can be specified for a single language.\n"+
		"Example:\n"+
		"  next \\\n"+
		"    -T go=./templates/go \\\n"+
		"    -T go=./templates/go_extra.npl \\\n"+
		"    -T python=./templates/python.npl\n",
	))

	// @api(CommandLine/Flag/-M) represents the language-specific type mappings and features.
	// **%T%**, **%T.E%**, **%N%**, **%K%**, **%V%** are placeholders replaced with actual types or values.
	// **%T.E%** is the final element type of a vector or array. It's used to get the element type of multi-dimensional arrays.
	//
	// Example:
	//
	//	```sh
	//	next -M cpp.vector="std::vector<%T%>" \
	//	     -M java.array="ArrayList<%T%>" \
	//	     -M go.map="map[%K%]%V%" \
	//	     -M python.ext=.py \
	//	     -M protobuf.vector="repeated %T.E%" \
	//	     -M ruby.comment="# %T%" \
	//	     ...
	//	```
	flagSet.Var(&o.Mapping, "M", u(""+
		"Configure language-specific type mappings and features.\n"+
		"`LANG.KEY=VALUE` specifies the mappings for a given language and type/feature.\n"+
		"Type mappings: Map Next types to language-specific types.\n"+
		"  Primitive types: int, int8, int16, int32, int64, bool, string, any, byte, bytes\n"+
		"  Generic types: vector, array, map\n"+
		"    "+b("%T%")+", "+b("%T.E%")+", "+b("%N%")+", "+b("%K%")+", "+b("%V%")+" are placeholders replaced with actual types or values.\n"+
		"    "+b("%T.E%")+" is the final element type of a vector or array. It's used to get the element type of multi-dimensional arrays.\n"+
		"Feature mappings: Set language-specific properties like file extensions or comment styles.\n"+
		"Example:\n"+
		"  next \\\n"+
		"    -M cpp.vector=\"std::vector<%T%>\" \\\n"+
		"    -M java.array=\"ArrayList<%T%>\" \\\n"+
		"    -M go.map=\"map[%K%]%V%\" \\\n"+
		"    -M python.ext=.py \\\n"+
		"    -M protobuf.vector=\"repeated %T.E%\" \\\n"+
		"    -M ruby.comment=\"# %T%\"\n",
	))

	// @api(CommandLine/Flag/-X) represents the custom annotation solver programs for code generation.
	// Annotation solvers are executed in a separate process to solve annotations.
	// All annotations are passed to the solver program via stdin and stdout.
	// The built-in annotation `next` is reserved for the Next compiler.
	//
	// Example:
	//
	//	```sh
	//	next -X message="message-type-allocator message-types.json" ...
	//	```
	//
	// :::tip
	//
	// In the example above, the `message-type-allocator` is a custom annotation solver program that
	// reads the message types from the `message-types.json` file and rewrite the message types to the
	// `message-types.json` file.
	//
	// :::
	flagSet.Var(&o.Solvers, "X", u(""+
		"Specify custom annotation solver programs for code generation.\n"+
		"`ANNOTATION=PROGRAM` defines the target annotation and its solver program.\n"+
		"Annotation solvers are executed in a separate process to solve annotations.\n"+
		"All annotations are passed to the solver program via stdin and stdout.\n"+
		b("NOTE")+": built-in annotation 'next' is reserved for the Next compiler.\n"+
		"Example:\n"+
		"  next -X message=\"message-type-allocator message-types.json\"\n",
	))
}

func (o *Configuration) resolvePath(currentPath string) {
	if o.Grammar != "" {
		o.Grammar = filepath.Join(currentPath, o.Grammar)
	}
	for k, v := range o.Output {
		o.Output[k] = filepath.Join(currentPath, v)
	}
	for k, v := range o.Templates {
		for i, t := range v {
			o.Templates[k][i] = filepath.Join(currentPath, t)
		}
	}
}

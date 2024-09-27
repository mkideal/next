# CommandLine {#user-content-CommandLine}
## Flag {#user-content-CommandLine_Flag}
### -D {#user-content-CommandLine_Flag_-D}

`-D` represents the custom environment variables for code generation. The value is a map of environment variable names and their optional values.

Example:

```sh
next -D VERSION=2.1 -D DEBUG -D NAME=myapp ...
```

```npl
{{env.NAME}}
{{env.VERSION}}
```

Output:

```
myapp
2.1
```

### -M {#user-content-CommandLine_Flag_-M}

`-M` represents the language-specific type mappings and features.

Example:

```sh
next -M cpp.vector="std::vector<%T%>" \
     -M java.array="ArrayList<%T%>" \
     -M go.map="map[%K%]%V%" \
     -M python.ext=.py \
     -M ruby.comment="# %T%" \
     ...
```

### -O {#user-content-CommandLine_Flag_-O}

`-O` represents the output directories for generated code of each target language.

Example:

```sh
next -O go=./output/go -O ts=./output/ts ...
```

:::tip

The `{{meta.path}}` is relative to the output directory.

:::

### -T {#user-content-CommandLine_Flag_-T}

`-T` represents the custom template directories or files for each target language. You can specify multiple templates for a single language.

Example:

```sh
next -T go=./templates/go \
     -T go=./templates/go_extra.npl \
     -T python=./templates/python.npl \
     ...
```

### -X {#user-content-CommandLine_Flag_-X}

`-X` represents the custom annotation solver programs for code generation. Annotation solvers are executed in a separate process to solve annotations. All annotations are passed to the solver program via stdin and stdout. The built-in annotation `next` is reserved for the Next compiler.

Example:

```sh
next -X message="message-type-allocator message-types.json" ...
```

:::tip

In the example above, the `message-type-allocator` is a custom annotation solver program that reads the message types from the `message-types.json` file and rewrite the message types to the `message-types.json` file.

:::

### -v {#user-content-CommandLine_Flag_-v}

`-v` represents the verbosity level of the compiler. The default value is **0**, which only shows error messages. The value **1** shows debug messages, and **2** shows trace messages. Usually, the trace message is used for debugging the compiler itself. Levels **1** (debug) and above enable execution of:
- `print` and `printf` in Next source files (.next).
- `debug` in Next template files (.npl).

Example:

```sh
next -v 1 ...
```

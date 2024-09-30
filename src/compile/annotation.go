package compile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/mattn/go-shellwords"

	"github.com/next/next/api"
	"github.com/next/next/src/ast"
	"github.com/next/next/src/constant"
	"github.com/next/next/src/token"
)

// @api(Object/Common/Annotations) represents a group of annotations by `name` => [Annotation](#Object/Common/Annotation).
//
// Annotations is a map that stores multiple annotations for a given entity.
// The key is the annotation name (string), and the value is the corresponding [Annotation](#Object/Common/Annotation) object.
type Annotations map[string]Annotation

func (a Annotations) get(name string) Annotation {
	if a == nil {
		return nil
	}
	return a[name]
}

// @api(Object/Common/Annotations.Contains) reports whether the annotations contain the given annotation.
//
// Example:
//
//	```next
//	@json(omitempty)
//	struct User {/*...*/}
//	```
//
//	```npl
//	{{if .Annotations.Contains "json"}}
//	{{/* do something */}}
//	{{end}}
//	```
func (a Annotations) Contains(name string) bool {
	if a == nil {
		return false
	}
	_, ok := a[name]
	return ok
}

// @api(Object/Common/Annotation) represents an annotation by `name` => value.
//
// Annotation is a map that stores the parameters of a single annotation.
// It allows for flexible parameter types, including strings, numbers, booleans and [Type](#Object/Common/Type)s.
//
// Example:
//
// Next code:
//
//	```next
//	@json(omitempty)
//	@event(name="Login")
//	@message(name="Login", type=100)
//	struct Login {}
//
//	@next(type=int8)
//	enum Color {
//		Red = 1;
//		Green = 2;
//		Blue = 3;
//	}
//	```
//
// Will be represented as:
//
//	```npl
//	{{- define "go/struct" -}}
//	{{.Annotations.json.omitempty}}
//	{{.Annotations.event.name}}
//	{{.Annotations.message.name}}
//	{{.Annotations.message.type}}
//	{{- end}}
//
//	{{- define "go/enum" -}}
//	{{.Annotations.next.type}}
//	{{- end}}
//	```
//
// Output:
//
//	```
//	true
//	Login
//	Login
//	100
//	int8
//	```
//
// The `next` annotation is used to pass information to the next compiler. It's a reserved
// annotation and should not be used for other purposes. The `next` annotation can be annotated
// to `package` statements, `const` declarations, `enum` declarations, `struct` declarations,
// `field` declarations, `interface` declarations, `method` declarations, and `parameter` declarations.
//
// :::note
//
// parameter names **MUST NOT** start with an uppercase letter, as this is reserved for the next compiler.
//
//	```next
//	@next(type=100) // OK
//	@next(_type=100) // OK
//
//	// This will error
//	@next(Type=100)
//	// invalid parameter name "Type": must not start with an uppercase letter (A-Z), e.g., "type"
//	```
//
// :::
type Annotation map[string]any

func (a Annotation) get(name string) any {
	if a == nil {
		return nil
	}
	return a[name]
}

// @api(Object/Common/Annotation.Contains) reports whether the annotation contains the given parameter.
//
// Example:
//
//	```next
//	@json(omitempty)
//	struct User {/*...*/}
//	```
//
//	```npl
//	{{if .Annotations.json.Contains "omitempty"}}
//	{{/* do something */}}
//	{{end}}
//	```
//
// :::note
//
// If you want to check whether the annotation has a non-empty value, you can use the parameter name directly.
//
//	```npl
//	{{if .Annotations.json.omitempty}}
//	{{/* do something */}}
//	{{end}}
//	```
//
// :::
func (a Annotation) Contains(name string) bool {
	if a == nil {
		return false
	}
	_, ok := a[name]
	return ok
}

const __pos__ = "__pos__"

// @api(Object/Common/Annotation.Pos) returns the position of the annotation in the source code.
// It's useful to provide a better error message when needed.
//
// Example:
//
//	```next title="example.next" showLineNumbers
//	package demo;
//
//	@message(type=100)
//	struct Login {/*...*/}
//	```
//
//	```npl
//	{{error "%s: Something went wrong" .Annotations.message.Pos}}
//	```
//
// Output:
//
//	```
//	example.next:3:1: Something went wrong
//	```
func (a Annotation) Pos() Position {
	return a.getPos(__pos__)
}

// @api(Object/Common/Annotation.NamePos) returns the position of the annotation name in the source code.
// It's useful to provide a better error message when needed.
//
// Example:
//
//	```next title="example.next" showLineNumbers
//	package demo;
//
//	@message(type=100)
//	struct Login {/*...*/}
//	```
//
//	```npl
//	{{error "%s: Something went wrong" (.Annotations.message.NamePos "type")}}
//	```
//
// Output:
//
//	```
//	example.next:3:10: Something went wrong
//	```
func (a Annotation) NamePos(name string) Position {
	return a.getPos(name)
}

// @api(Object/Common/Annotation.ValuePos) returns the position of the annotation value in the source code.
// It's useful to provide a better error message when needed.
//
// Example:
//
//	```next title="example.next" showLineNumbers
//	package demo;
//
//	@message(type=100)
//	struct Login {/*...*/}
//	```
//
//	```npl
//	{{error "%s: Something went wrong" (.Annotations.message.ValuePos "type")}}
//	```
//
// Output:
//
//	```
//	example.next:3:15: Something went wrong
//	```
func (a Annotation) ValuePos(name string) Position {
	return a.getPos(name + ":value")
}

func (a Annotation) getPos(key string) Position {
	if a == nil {
		return Position{}
	}
	positions, ok := a[__pos__].(map[string]Position)
	if !ok {
		return Position{}
	}
	return positions[key]
}

func (a Annotation) setPos(key string, pos Position) {
	positions, ok := a[__pos__].(map[string]Position)
	if !ok {
		positions = make(map[string]Position)
		a[__pos__] = positions
	}
	positions[key] = pos
}

// @api(Object/Common/Annotation/decl.available)
// The `@next(available="expression")` annotation for `file`, `const`, `enum`, `struct`, `field`, `interface`, `method`
// availability of the declaration. The `expression` is a boolean expression that can be used to control the availability
// of the declaration in the target language. Supported operators are `&`, `|`, `!`, `(`, `)`, and `true`, `false`.
//
// Example:
//
//	```next
//	@next(available="c|cpp|java|go|csharp")
//	struct Point {
//		int x;
//		int y;
//
//		@next(available="c | cpp | go")
//		int z;
//
//		@next(available="!java & !c")
//		int w;
//	}
//	```

// @api(Object/Common/Annotation/package)
// The `next` annotation for `package` statements used to control the package behavior for specific
// languages. The `next` annotation can be used to set the package name, package path, and some other
// package-related information.
//
// For any language `L`, the `next` annotation for `package` statements is defined as `@next(L_package="package_info")`.
//
// Example:
//
//	```next
//	@next(
//		c_package="DEMO_",
//		cpp_package="demo",
//		java_package="com.exmaple.demo",
//		go_package="github.com/next/demo",
//		csharp_package="demo",
//	)
//	```
//
//	```npl
//	{{.Package.Annotations.next.c_package}}
//	{{.Package.Annotations.next.cpp_package}}
//	{{.Package.Annotations.next.java_package}}
//	{{.Package.Annotations.next.go_package}}
//	{{.Package.Annotations.next.csharp_package}}
//	```
//
// There are some reserved keys for the `next` annotation for `package` statements.

// @api(Object/Common/Annotation/package.go_imports) represents a list of import paths for Go packages,
// separated by commas: `@next(go_imports="fmt.Printf,*io.Reader")`.
//
// :::note
//
// **`*`** is required to import types.
//
// :::
//
// Example:
//
//	```next
//	@next(go_imports="fmt.Printf,*io.Reader")
//	package demo;
//	```

// @api(Object/Common/Annotation/enum)
// The `next` annotation for `enum` declarations used to control the enum behavior.

// @api(Object/Common/Annotation/enum.type) specifies the underlying type of the enum.
//
// Example:
//
//	```next
//	@next(type=int8)
//	enum Color {
//		Red = 1;
//		Green = 2;
//		Blue = 3;
//	}
//	```
//
// Output in Go:
//
//	```go
//	type Color int8
//
//	const (
//		ColorRed Color = 1
//		ColorGreen Color = 2
//		ColorBlue Color = 3
//	)
//	```
//
// Output in C++:
//
//	```cpp
//	enum class Color : int8_t {
//		Red = 1,
//		Green = 2,
//		Blue = 3,
//	};
//	```

// @api(Object/Common/Annotation/struct)
// The `next` annotation for `struct` declarations used to control the struct behavior.
// `L_alias` is a alias for the struct name in language `L`. It's used to reference an external type
// in the target language.
//
// Example:
//
//	```next
//	@next(rust_alias="u128")
//	struct uint128 {
//		int64 low;
//		int64 high;
//	}
//
//	@next(go_alias="complex128")
//	struct Complex {
//		float64 real;
//		float64 imag;
//	}
//
//	struct Contract {
//		uint128 address;
//		Complex complex;
//	}
//	```
//
// This will don't generate the `uint128` struct in the `rust` language, but use `u128` instead.
// And in the `go` language, it will use `complex128` instead of `Complex`.

// @api(Object/Common/Annotation/interface)
// The `next` annotation for `interface` declarations used to control the interface behavior.
// `L_alias` is a alias for the interface name in language `L`. It's used to reference an external type
// in the target language.
//
// Example:
//
//	```next
//	@next(
//		available="go|java",
//		go_alias="net/http.Handler",
//		java_alias="java.util.function.Function<com.sun.net.httpserver.HttpExchange, String>",
//	)
//	interface HTTPHandler {}
//
//	@next(available="go|java")
//	interface HTTPServer {
//		@next(error)
//		Handle(string path, HTTPHandler handler);
//	}
//	```

// @api(Object/Common/Annotation/method)
// The `next` annotation for `method` declarations used to control the method behavior.

// @api(Object/Common/Annotation/method.error)
// The `@next(error)` annotation used to indicate that the method returns an error or throws an exception.
//
// Example:
//
//	```next
//	interface Parser {
//		@next(error)
//		parse(string s) int;
//	}
//	```
//
// Output in Go:
//
//	```go
//	type Parser interface {
//		Parse(s string) (int, error)
//	}
//	```
//
// Output in C++:
//
//	```cpp
//	class Parser {
//	public:
//		int parse(const std::string& s) const;
//	};
//	```
//
// Output in Java:
//
//	```java
//	interface Parser {
//		int parse(String s) throws Exception;
//	}
//	```

// @api(Object/Common/Annotation/method.mut)
// The `@next(mut)` annotation used to indicate that the method is a mutable method, which means
// it can modify the object's state.
//
// Example:
//
//	```next
//	interface Writer {
//		@next(error, mut)
//		write(string data);
//	}
//	```
//
// Output in Go:
//
//	```go
//	type Writer interface {
//		Write(data string) error
//	}
//	```
//
// Output in C++:
//
//	```cpp
//	class Writer {
//	public:
//		void write(const std::string& data);
//	};
//	```

// @api(Object/Common/Annotation/param)
// The `next` annotation for `parameter` declarations used to control the parameter behavior.

// @api(Object/Common/Annotation/param.mut)
// The `@next(mut)` annotation used to indicate that the parameter is mutable.
//
// Example:
//
//	```next
//	interface Reader {
//		@next(error);
//		read(@next(mut) string data);
//	}
//	```
//
// Output in Go:
//
//	```go
//	type Reader interface {
//		Read(data string) error
//	}
//	```
//
// Output in C++:
//
//	```cpp
//	class Reader {
//	public:
//		void read(std::string& data);
//	};
//	```

// linkedAnnotation represents an annotation linked to a declaration.
type linkedAnnotation struct {
	name       string
	annotation Annotation

	// obj is declared as type `Object` to avoid circular dependency.
	// Actually, it's a `Node`.
	obj Object
}

// resolveAnnotations resolves an annotation group
func resolveAnnotations(c *Compiler, file *File, obj Node, annotations *ast.AnnotationGroup) Annotations {
	if annotations == nil {
		return nil
	}
	result := make(Annotations)
	for _, a := range annotations.List {
		if _, dup := result[a.Name.Name]; dup {
			c.addErrorf(a.Pos(), "annotation %s redeclared", a.Name.Name)
			continue
		}
		annotation := make(Annotation)
		la := &linkedAnnotation{
			name:       a.Name.Name,
			obj:        obj,
			annotation: annotation,
		}
		pos := obj.Pos().pos
		c.annotations[pos] = la
		annotation.setPos(__pos__, positionFor(c, a.Pos()))
		for _, p := range a.Params {
			name := p.Name.Name
			if name == "" {
				c.addErrorf(p.Name.Pos(), "parameter name cannot be empty")
				continue
			}
			if name[0] >= 'A' && name[0] <= 'Z' {
				c.addErrorf(p.Name.Pos(), "invalid parameter name %q: must not start with an uppercase letter (A-Z), e.g., %q", name, strings.ToLower(name[:1])+name[1:])
				continue
			}

			if _, dup := annotation[name]; dup {
				c.addErrorf(p.Pos(), "parameter %q redefined, previous definition at %s", name, annotation.NamePos(name))
				continue
			}
			var value any
			if p.Value == nil {
				value = true
			} else if t, ok := p.Value.(ast.Type); ok {
				if typ := c.resolveType(file, t, true); typ != nil {
					value = typ
				}
			}
			if value == nil {
				if e, ok := p.Value.(ast.Expr); ok {
					value = constant.Underlying(c.resolveValue(file, e, nil))
				}
			}
			if value != nil {
				annotation[name] = value
				annotation.setPos(p.Name.Name, positionFor(c, p.Name.NamePos))
				if p.Value != nil {
					annotation.setPos(p.Name.Name+":value", positionFor(c, p.Value.Pos()))
				}
			} else {
				c.addErrorf(p.Pos(), "unexpected parameter %T", p.Value)
			}
		}
		result[a.Name.Name] = annotation
	}
	return result
}

func (c *Compiler) solveAnnotations() error {
	parser := shellwords.NewParser()
	parser.ParseEnv = true
	programs := make(map[string][][]string)
	keys := make([]string, 0, len(c.options.Solvers))
	for name, solvers := range c.options.Solvers {
		if name == "next" {
			return fmt.Errorf("'next' is a reserved annotation for the next compiler, please don't use solvers for it")
		}
		keys = append(keys, name)
		for _, solver := range solvers {
			words, err := parser.Parse(solver)
			if err != nil {
				return err
			}
			if len(words) == 0 {
				return fmt.Errorf("empty solver for %q", name)
			}
			programs[name] = append(programs[name], words)
		}
	}
	sort.Strings(keys)
	for _, name := range keys {
		for _, words := range programs[name] {
			req := c.createAnnotationSolverRequest(name)
			if len(req.Annotations) == 0 {
				continue
			}
			c.Trace("run solver %q", words)
			cmd := exec.Command(words[0], words[1:]...)
			var stdin bytes.Buffer
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdin = &stdin
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := json.NewEncoder(&stdin).Encode(req); err != nil {
				return fmt.Errorf("failed to encode request for solver %q: %v", name, err)
			}
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to run solver %q: %v\n%s", name, err, stderr.String())
			}
			var res api.AnnotationSolverResponse
			if err := json.NewDecoder(&stdout).Decode(&res); err != nil {
				return fmt.Errorf("failed to decode response for solver %q: %v", name, err)
			}
			for id, a := range res.Annotations {
				la, ok := c.annotations[token.Pos(id)]
				if !ok {
					return fmt.Errorf("solver %q: annotation %d not found", words, id)
				}
				for name, param := range a.Params {
					var v constant.Value
					if param.Value != nil {
						v = constant.Make(param.Value)
						if v.Kind() == constant.Unknown {
							return fmt.Errorf("solver %q: invalid value for parameter %q in annotation %q: %v", words, name, la.name, param.Value)
						}
						c.Trace("solver %q: set %q.%q to %v", words, la.name, name, param.Value)
					}
					_, ok := la.annotation[name]
					if !ok {
						if v == nil {
							return fmt.Errorf("solver %q: add an invalid value for parameter %q in annotation %q: %v", words, name, la.name, param.Value)
						}
						la.annotation[name] = constant.Underlying(v)
						continue
					}
					if v != nil {
						la.annotation[name] = constant.Underlying(v)
					} else {
						c.Trace("solver %q: remove %q.%q", words, la.name, name)
						delete(la.annotation, name)
					}
				}
			}
		}
	}
	return nil
}

func (c *Compiler) createAnnotationSolverRequest(name string) *api.AnnotationSolverRequest {
	req := &api.AnnotationSolverRequest{
		Objects:     make(map[api.ID]*api.Object),
		Annotations: make(map[api.ID]*api.Annotation),
	}
	var objects = make(map[token.Pos]Node)
	for pos, a := range c.annotations {
		if a.name != name {
			continue
		}
		var params = make(map[string]api.Parameter)
		for name, param := range a.annotation {
			params[name] = api.Parameter{
				Name:  name,
				Value: param,
			}
		}
		// a.obj is declared as type `Object` to avoid circular dependency.
		// Actually, it's a `Node`.
		obj := a.obj.(Node)
		req.Annotations[api.ID(pos)] = &api.Annotation{
			ID:     api.ID(pos),
			Object: api.ID(obj.Pos().pos),
			Params: params,
		}
		objects[obj.Pos().pos] = obj
	}
	for pos, obj := range objects {
		req.Objects[api.ID(pos)] = &api.Object{
			ID:   api.ID(pos),
			Type: obj.Typeof(),
			Name: obj.Name(),
			Pkg:  obj.File().pkg.name,
			File: obj.File().Path,
		}
	}
	return req
}

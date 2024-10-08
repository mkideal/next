//go:generate go run github.com/gopherd/tools/cmd/docgen@v0.0.8 -ext .mdx -I ./ -o ../../website/docs/api/preview -level 0 -M "---" -M "pagination_prev: null" -M "pagination_next: null" -M "---"
package grammar

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/gopherd/core/text/templates"
	"github.com/next/next/src/internal/stringutil"
)

const (
	Bool   = "bool"
	Int    = "int"
	Float  = "float"
	String = "string"
	Type   = "type"
	Any    = "any"
)

func types(s ...string) []string {
	return s
}

var validAnnotationParameterTypes = []string{Bool, Int, Float, String, Type, Any}
var validConstTypes = []string{Bool, Int, Float, String}
var validEnumMemberTypes = []string{Int, Float, String}

func duplicated[S ~[]T, T comparable](s S) error {
	if len(s) < 2 {
		return nil
	}
	seen := make(map[T]struct{}, len(s))
	for _, v := range s {
		if _, ok := seen[v]; ok {
			return fmt.Errorf("duplicated %v", v)
		}
		seen[v] = struct{}{}
	}
	return nil
}

// Context represents the contextual data by the key-value pair.
// The key is a string and the value is a JSON object.
type Context struct {
	Annotations          map[string]json.RawMessage `json:"annotations"`
	AnnotationParameters map[string]json.RawMessage `json:"annotation_parameters"`
	Validators           map[string]json.RawMessage `json:"validators"`
}

// Options represents the options with the id and/or the options.
// If the id is empty, the options are used. Otherwise, the id is used.
// The options are resolved by the context using the id as the key.
type Options[T any] struct {
	id    string
	value T
}

func opt[T any](options T) Options[T] {
	return Options[T]{value: options}
}

func (o Options[T]) Value() T {
	return o.value
}

func (o Options[T]) MarshalJSON() ([]byte, error) {
	if o.id == "" {
		return json.Marshal(o.value)
	}
	return json.Marshal(o.id)
}

func (o *Options[T]) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '"' {
		return json.Unmarshal(data, &o.id)
	}
	return json.Unmarshal(data, &o.value)
}

func (o *Options[T]) resolve(sourceName string, source map[string]json.RawMessage) error {
	if o.id == "" {
		return nil
	}
	if source == nil {
		return fmt.Errorf("can't resolve %q because of %s not defined", o.id, sourceName)
	}
	v, ok := source[o.id]
	if !ok {
		return fmt.Errorf("%q not found in %s", o.id, sourceName)
	}
	if err := json.Unmarshal(v, &o.value); err != nil {
		return fmt.Errorf("failed to decode %q in %s: %s", o.id, sourceName, strings.TrimPrefix(err.Error(), "json: "))
	}
	return nil
}

// @api(Grammar) represents the custom grammar for the next files.
//
//	import CodeBlock from "@theme/CodeBlock";
//	import Tabs from "@theme/Tabs";
//	import TabItem from "@theme/TabItem";
//	import ExampleGrammarYAMLSource from "!!raw-loader!@site/example/grammar.yaml";
//	import ExampleGrammarJSONSource from "!!raw-loader!@site/example/grammar.json";
//
// The grammar is used to define a subset of the next files. It can limit the features of the next code according
// by your requirements. The grammar is a yaml file that contains rules.
//
// Here is an example of the grammar file:
//
//	<Tabs
//		defaultValue="yaml"
//		values={[
//			{label: 'YAML', value: 'yaml'},
//			{label: 'JSON', value: 'json'},
//		]}
//	>
//	<TabItem value="yaml">
//		<CodeBlock language="yaml" title="grammar.yaml">
//			{ExampleGrammarYAMLSource}
//		</CodeBlock>
//	</TabItem>
//	<TabItem value="json">
//		<CodeBlock language="json" title="grammar.json">
//			{ExampleGrammarJSONSource}
//		</CodeBlock>
//	</TabItem>
//	</Tabs>
//
// It extends **built-in** grammar and defines a `@message` annotation for `struct` objects. For example:
//
//	```next
//	package demo;
//
//	@message(type=1, req)
//	struct LoginRequest {/*...*/}
//
//	@message(type=2)
//	struct LoginResponse {/*...*/}
//	```
//
//	:::tip
//	Run the following command to show the **built-in** grammar:
//
//	```sh
//	next grammar
//	# Or outputs the grammar to a file
//	next grammar grammar.yaml
//	next grammar grammar.json
//	```
//	:::
type Grammar struct {
	builtin bool `json:"-"`

	Context   Context   `json:"context"`
	Package   Package   `json:"package"`
	Import    Import    `json:"import"`
	Const     Const     `json:"const"`
	Enum      Enum      `json:"enum"`
	Struct    Struct    `json:"struct"`
	Interface Interface `json:"interface"`
}

type Validators []Options[Validator]

// @api(Grammar/Common/Validator) represents the validator for the grammar rules.
type Validator struct {
	// @api(Grammar/Common/Validator.Name) represents the validator name.
	Name string `json:"name"`

	// @api(Grammar/Common/Validator.Expression) represents the validator expression.
	// The expression is a template string that can access the data by the `.` operator. The expression
	// must return a boolean value.
	//
	// The data is the current context object. For example, **package** object for the package validator.
	Expression string `json:"expression"`

	// @api(Grammar/Common/Validator.Message) represents the error message when the validator is failed.
	Message string `json:"message"`
}

func (v Validator) Validate(data any) (bool, error) {
	result, err := templates.Execute(v.Name, v.Expression, data)
	if err != nil {
		return false, fmt.Errorf("failed to execute validator %q: %w", v.Name, err)
	}
	ok, err := strconv.ParseBool(result)
	if err != nil {
		return false, fmt.Errorf("validator %q must return a boolean value: %w", v.Name, err)
	}
	return ok, nil
}

// @api(Grammar/Common/Annotation) represents the annotation grammar rules.
// Only the built-in annotations are supported if no annotations are defined in the grammar.
//
// Example:
//
//	<Tabs
//		defaultValue="yaml"
//		values={[
//			{label: 'YAML', value: 'yaml'},
//			{label: 'JSON', value: 'json'},
//		]}
//	>
//	<TabItem value="json">
//	```json
//	{
//	  "struct": {
//	    "annotations": [
//	      {
//	        "name": "message",
//	        "description": "Sets the struct as a message.",
//	        "parameters": [
//	          {
//	            "name": "type",
//	            "description": "Sets the message type id.",
//	            "type": "int",
//	            "required": true
//	            "validators": [
//	              {
//	                "name": "MessageTypeMustBePositive",
//	                "expression": "{{gt . 0}}",
//	                "message": "message type must be positive"
//	              }
//	            ]
//	          }
//	        ]
//	      }
//	    ]
//	  }
//	}
//	```
//	</TabItem>
//	<TabItem value="yaml">
//	```yaml
//	struct:
//	  annotations:
//	    - name: message
//	      description: Sets the struct as a message.
//	      parameters:
//	        - name: type
//	          description: Sets the message type id.
//	          types: [int]
//	          required: true
//	          validators:
//	            - name: MessageTypeMustBePositive
//	              expression: "{{gt . 0}}"
//	              message: message type must be positive
//	```
//	</TabItem>
//	</Tabs>
//
//	```next
//	package demo;
//
//	// Good
//	@message(type=1)
//	struct User {
//	  int id;
//	  string name;
//	}
//
//	// This will error
//	@message
//	struct User {
//	  int id;
//	  string name;
//	}
//	// Error: message type is required
//
//	// This will error
//	@message(type)
//	struct User {
//	  int id;
//	  string name;
//	}
//	// Error: message type must be an integer
//
//	// This will error
//	@message(type=-1)
//	struct User {
//	  int id;
//	  string name;
//	}
//	// Error: message type must be positive
//	```
type Annotation struct {
	// @api(Grammar/Common/Annotation.Name) represents the annotation name.
	Name string `json:"name"`

	// @api(Grammar/Common/Annotation.Description) represents the annotation description.
	Description string `json:"description"`

	// @api(Grammar/Common/Annotation.Parameters) represents the annotation parameters.
	Parameters []Options[AnnotationParameter] `json:"parameters"`
}

func LookupAnnotation(annotations Annotations, name string) *Annotation {
	for i := range annotations {
		a := &annotations[i].value
		if a.Name == name {
			return a
		}
	}
	return nil
}

// @api(Grammar/Common/Annotations) represents a list of annotations.
type Annotations []Options[Annotation]

// @api(Grammar/Common/AnnotationParameter) represents the annotation parameter grammar rules.
// If no parameters are defined, the annotation does not have any parameters.
type AnnotationParameter struct {
	// @api(Grammar/Common/AnnotationParameter.Name) represents the parameter name pattern.
	// The name pattern is a regular expression that can be used to match the annotation parameter name.
	//
	// Example:
	//
	// - "**type**": matches the annotation name `type`
	// - "**x|y**": matches the annotation name `x` or `y`
	// - "**.+_package**": matches the annotation name that ends with `_package`, for example, `cpp_package`, `java_package`, etc.
	Name string `json:"name"`

	// @api(Grammar/Common/AnnotationParameter.Description) represents the parameter description.
	Description string `json:"description"`

	// @api(Grammar/Common/AnnotationParameter.Types) represents the parameter type.
	// The type is a string that can be one of the following types:
	//
	// - **bool**: boolean type, the value can be `true` or `false`
	// - **int**: integer type, the value can be a positive or negative integer, for example, `123`
	// - **float**: float type, the value can be a positive or negative float, for example, `1.23`
	// - **string**: string type, the value can be a string, for example, `"hello"`
	// - **type**: any type name, for example, `int`, `float`, `string`, etc. Custom type names are supported.
	Types []string `json:"types"`

	// @api(Grammar/Common/AnnotationParameter.Required) represents the parameter is required or not.
	Required bool `json:"required"`

	// @api(Grammar/Common/AnnotationParameter.Validators) represents the [Validator](#Grammar/Common/Validator) for the annotation parameter.
	Validators Validators `json:"validators"`

	parsed struct {
		name *regexp.Regexp
	} `json:"-"`
}

func LookupAnnotationParameter(parameters []Options[AnnotationParameter], name string) *AnnotationParameter {
	for i := range parameters {
		p := &parameters[i].value
		if p.parsed.name != nil {
			if p.parsed.name.MatchString(name) {
				return p
			}
		} else if p.Name == name {
			return p
		}
	}
	return nil
}

// @api(Grammar/Package) represents the grammar rules for the package declaration.
type Package struct {
	// @api(Grammar/Package.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the package declaration.
	Annotations Annotations `json:"annotations"`

	// @api(Grammar/Package.Validators) represents the [Validator](#Grammar/Common/Validator) for the package declaration.
	// It's used to validate the package name. For example, You can limit the package name must be
	// not start with a "_" character. The validator expression is a template string that can access
	// package name by `.Name`.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "package": {
	//	    "validators": [
	//	      {
	//	        "name": "PackageNameNotStartWithUnderscore",
	//	        "expression": "{{not (hasPrefix `_` .Name)}}",
	//	        "message": "package name must not start with an underscore"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	package:
	//	  validators:
	//	    - name: PackageNameNotStartWithUnderscore
	//	      expression: "{{not (hasPrefix `_` .Name)}}"
	//	      message: package name must not start with an underscore
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	// This will error
	//	package _test;
	//	// Error: package name must not start with an underscore
	//	```
	Validators Validators `json:"validators"`
}

// @api(Grammar/Import) represents the grammar rules for the import declaration.
type Import struct {
	// @api(Grammar/Import.Disabled) represents the import declaration is off or not.
	// If the import declaration is off, the import declaration is not allowed in the next files.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "import": {
	//	    "disabled": true
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	import:
	//	  disabled: true
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// This will error
	//	import "other.next";
	//	// Error: import declaration is not allowed
	//	```
	Disabled bool `json:"disabled"`
}

// @api(Grammar/Const) represents the grammar rules for the const declaration.
type Const struct {
	// @api(Grammar/Const.Disabled) represents the const declaration is off or not.
	// If the const declaration is off, the const declaration is not allowed in the next files.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "const": {
	//	    "disabled": true
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	const:
	//	  disabled: true
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// This will error
	//	const x = 1;
	//	// Error: const declaration is not allowed
	//	```
	Disabled bool `json:"disabled"`

	// @api(Grammar/Const.Types) represents a list of type names that are supported in the const declaration.
	// If no types are defined, the const declaration supports all types. Otherwise, the const declaration
	// only supports the specified types.
	//
	// Currently, each type name can be one of the following types:
	//
	// - **bool**: boolean type, the value can be `true` or `false`
	// - **int**: integer type, the value can be a positive or negative integer, for example, `123`
	// - **float**: float type, the value can be a positive or negative float, for example, `1.23`
	// - **string**: string type, the value can be a string, for example, `"hello"`
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "const": {
	//	    "types": ["int", "float"]
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	const:
	//	  types:
	//	    - int
	//	    - float
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	const x = 1; // Good
	//	const y = 1.23; // Good
	//
	//	// This will error
	//	const z = "hello";
	//	// Error: string type is not allowed in the const declaration
	//	```
	Types []string `json:"types"`

	// @api(Grammar/Const.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the const declaration.
	Annotations Annotations `json:"annotations"`

	// @api(Grammar/Const.Validators) represents the [Validator](#Grammar/Common/Validator) for the const declaration.
	// It's used to validate the const name. You can access the const name by `.Name`.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "const": {
	//	    "validators": [
	//	      {
	//	        "name": "ConstNameMustBeCapitalized",
	//	        "expression": "{{eq .Name (.Name | capitalize)}}",
	//	        "message": "const name must be capitalized"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	const:
	//	  validators:
	//	    - name: ConstNameMustBeCapitalized
	//	      expression: "{{eq .Name (.Name | capitalize)}}"
	//	      message: const name must be capitalized
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	const Hello = 1; // Good
	//
	//	// This will error
	//	const world = 1;
	//	// Error: const name must be capitalized, expected: World
	//	```
	Validators Validators `json:"validators"`
}

// @api(Grammar/Enum) represents the grammar rules for the enum declaration.
type Enum struct {
	// @api(Grammar/Enum.Disabled) represents the enum declaration is off or not.
	// If the enum declaration is off, the enum declaration is not allowed in the next files.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "enum": {
	//	    "disabled": true
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	enum:
	//	  disabled: true
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// This will error
	//	enum Color {
	//	  Red;
	//	  Green;
	//	  Blue;
	//	}
	//	// Error: enum declaration is not allowed
	//	```
	Disabled bool `json:"disabled"`

	// @api(Grammar/Enum.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the enum declaration.
	Annotations Annotations `json:"annotations"`

	// @api(Grammar/Enum.Validators) represents the [Validator](#Grammar/Common/Validator) for the enum declaration.
	// It's used to validate the enum name. You can access the enum name by `.Name`.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "enum": {
	//	    "validators": [
	//	      {
	//	        "name": "EnumNameMustBeCapitalized",
	//	        "expression": "{{eq .Name (.Name | capitalize)}}",
	//	        "message": "enum name must be capitalized"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	enum:
	//	  validators:
	//	    - name: EnumNameMustBeCapitalized
	//	      expression: "{{eq .Name (.Name | capitalize)}}"
	//	      message: enum name must be capitalized
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// Good
	//	enum Color {
	//	  Red;
	//	  Green;
	//	  Blue;
	//	}
	//
	//	// This will error
	//	enum size {
	//	  Small;
	//	  Medium;
	//	  Large;
	//	}
	//	// Error: enum name must be capitalized, expected: Size
	//	```
	Validators Validators `json:"validators"`

	// @api(Grammar/Enum.Member) represents the [EnumMember](#Grammar/EnumMember) grammar rules for the enum declaration.
	Member EnumMember `json:"member"`
}

// @api(Grammar/EnumMember) represents the grammar rules for the enum member declaration.
type EnumMember struct {
	// @api(Grammar/EnumMember.Types) represents a list of type names that are supported in the enum declaration.
	//
	// If no types are defined, the enum declaration supports all types. Otherwise, the enum declaration
	// only supports the specified types.
	//
	// Currently, each type name can be one of the following types:
	//
	// - **int**: integer type, the value can be a positive or negative integer, for example, `123`
	// - **float**: float type, the value can be a positive or negative float, for example, `1.23`
	// - **string**: string type, the value can be a string, for example, `"hello"`
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "enum": {
	//	    "member": {
	//	      "types": ["int"]
	//	    }
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	enum:
	//	  member:
	//	    types:
	//	      - int
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// Good
	//	enum Color {
	//	  Red = 1,
	//	  Green = 2,
	//	  Blue = 3
	//	}
	//
	//	// This will error
	//	enum Size {
	//	  Small = "small",
	//	  Medium = "medium",
	//	  Large = "large"
	//	}
	//	// Error: string type is not allowed in the enum declaration
	//	```
	Types []string `json:"types"`

	// @api(Grammar/EnumMember.ValueRequired) represents the enum member value is required or not.
	// If the enum member value is required, the enum member value must be specified in the next files.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "enum": {
	//	    "member": {
	//	      "value_required": true
	//	    }
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	enum:
	//	  member:
	//	    value_required: true
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	enum Size {
	//	  Small = 1,
	//	  Medium = 2;
	//	  // This will error
	//	  Large;
	//	  // Error: enum member value is required
	//	}
	//	```
	ValueRequired bool `json:"value_required"`

	// @api(Grammar/EnumMember.ZeroRequired) represents the enum member zero value for integer types is required or not.
	//
	// If the enum member zero value is required, the enum member zero value must be specified in the next files.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "enum": {
	//	    "member": {
	//	      "zero_required": true
	//	    }
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	enum:
	//	  member:
	//	    zero_required: true
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// This will error
	//	enum Size {
	//	  Small = 1,
	//	  Medium = 2,
	//	  Large = 3;
	//	}
	//	// Error: enum member zero value is required, for example:
	//	// enum Size {
	//	//   Small = 0,
	//	//   Medium = 1,
	//	//   Large = 2
	//	// }
	//	```
	ZeroRequired bool `json:"zero_required"`

	// @api(Grammar/EnumMember.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the enum member declaration.
	Annotations Annotations `json:"annotations"`

	// @api(Grammar/EnumMember.Validators) represents the [Validator](#Grammar/Common/Validator) for the enum member declaration.
	//
	// It's used to validate the enum member name. You can access the enum member name by `.Name`.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "enum": {
	//	    "member": {
	//	      "validators": [
	//	        {
	//	          "name": "EnumMemberNameMustBeCapitalized",
	//	          "expression": "{{eq .Name (.Name | capitalize)}}",
	//	          "message": "enum member name must be capitalized"
	//	        }
	//	      ]
	//	    }
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	enum:
	//	  member:
	//	    validators:
	//	      - name: EnumMemberNameMustBeCapitalized
	//	        expression: "{{eq .Name (.Name | capitalize)}}"
	//	        message: enum member name must be capitalized
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	enum Size {
	//	  Small = 1,
	//	  Medium = 2,
	//	  // This will error
	//	  large = 3;
	//	  // Error: enum member name must be capitalized, expected: Large
	//	}
	//	```
	Validators Validators `json:"validators"`
}

// @api(Grammar/Struct) represents the grammar rules for the struct declaration.
type Struct struct {
	// @api(Grammar/Struct.Disabled) represents the struct declaration is off or not.
	// If the struct declaration is off, the struct declaration is not allowed in the next files.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "struct": {
	//	    "disabled": true
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	struct:
	//	  disabled: true
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// This will error
	//	struct User {
	//	  int id;
	//	  string name;
	//	}
	//	// Error: struct declaration is not allowed
	//	```
	Disabled bool `json:"disabled"`

	// @api(Grammar/Struct.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the struct declaration.
	Annotations Annotations `json:"annotations"`

	// @api(Grammar/Struct.Validators) represents the [Validator](#Grammar/Common/Validator) for the struct declaration.
	// It's used to validate the struct name. You can access the struct name by `.Name`.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "struct": {
	//	    "validators": [
	//	      {
	//	        "name": "StructNameMustBeCapitalized",
	//	        "expression": "{{eq .Name (.Name | capitalize)}}",
	//	        "message": "struct name must be capitalized"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	struct:
	//	  validators:
	//	    - name: StructNameMustBeCapitalized
	//	      expression: "{{eq .Name (.Name | capitalize)}}"
	//	      message: struct name must be capitalized
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// Good
	//	struct User {
	//	  int id;
	//	  string name;
	//	}
	//
	//	// This will error
	//	struct point {
	//	  int x;
	//	  int y;
	//	}
	//	// Error: struct name must be capitalized, expected: Point
	//	```
	Validators Validators `json:"validators"`

	// @api(Grammar/Struct.Field) represents the [StructField](#Grammar/StructField) grammar rules for the struct declaration.
	Field StructField `json:"field"`
}

// @api(Grammar/StructField) represents the grammar rules for the struct field declaration.
type StructField struct {
	// @api(Grammar/StructField.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the struct field declaration.
	Annotations Annotations `json:"annotations"`

	// @api(Grammar/StructField.Validators) represents the [Validator](#Grammar/Common/Validator) for the struct field declaration.
	// It's used to validate the struct field name. You can access the struct field name by `.Name`.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "struct": {
	//	    "field": {
	//	      "validators": [
	//	        {
	//	          "name": "StructFieldNameMustNotBeCapitalized",
	//	          "expression": "{{ne .Name (capitalize .Name)}}",
	//	          "message": "struct field name must not be capitalized"
	//	        }
	//	      ]
	//	    }
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	struct:
	//	  field:
	//	    validators:
	//	      - name: StructFieldNameMustNotBeCapitalized
	//	        expression: "{{ne .Name (capitalize .Name)}}"
	//	        message: struct field name must not be capitalized
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	struct User {
	//	  int id;
	//	  // This will error
	//	  string Name;
	//	  // Error: struct field name must not be capitalized, expected: name
	//	}
	//	```
	Validators Validators `json:"validators"`
}

// @api(Grammar/Interface) represents the grammar rules for the interface declaration.
type Interface struct {
	// @api(Grammar/Interface.Disabled) represents the interface declaration is off or not.
	// If the interface declaration is off, the interface declaration is not allowed in the next files.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "interface": {
	//	    "disabled": true
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	interface:
	//	  disabled: true
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// This will error
	//	interface User {
	//	  GetID() int;
	//	}
	//	// Error: interface declaration is not allowed
	//	```
	Disabled bool `json:"disabled"`

	// @api(Grammar/Interface.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the interface declaration.
	Annotations Annotations `json:"annotations"`

	// @api(Grammar/Interface.Validators) represents the [Validator](#Grammar/Common/Validator) for the interface declaration.
	// It's used to validate the interface name. You can access the interface name by `.Name`.
	//
	// Example:
	//
	//	<Tabs
	//		defaultValue="yaml"
	//		values={[
	//			{label: 'YAML', value: 'yaml'},
	//			{label: 'JSON', value: 'json'},
	//		]}
	//	>
	//	<TabItem value="json">
	//	```json
	//	{
	//	  "interface": {
	//	    "validators": [
	//	      {
	//	        "name": "InterfaceNameMustBeCapitalized",
	//	        "expression": "{{eq .Name (.Name | capitalize)}}",
	//	        "message": "interface name must be capitalized"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
	//	</TabItem>
	//	<TabItem value="yaml">
	//	```yaml
	//	interface:
	//	  validators:
	//	    - name: InterfaceNameMustBeCapitalized
	//	      expression: "{{eq .Name (.Name | capitalize)}}"
	//	      message: interface name must be capitalized
	//	```
	//	</TabItem>
	//	</Tabs>
	//
	//	```next
	//	package demo;
	//
	//	// Good
	//	interface User {
	//	  GetID() int;
	//	}
	//
	//	// This will error
	//	interface user {
	//	  GetName() string;
	//	}
	//	// Error: interface name must be capitalized, expected: User
	//	```
	Validators Validators `json:"validators"`

	// @api(Grammar/Interface.Method) represents the [InterfaceMethod](#Grammar/InterfaceMethod) grammar rules for the interface declaration.
	Method InterfaceMethod `json:"method"`
}

// @api(Grammar/InterfaceMethod) represents the grammar rules for the interface method declaration.
//
// Example:
//
//	<Tabs
//		defaultValue="yaml"
//		values={[
//			{label: 'YAML', value: 'yaml'},
//			{label: 'JSON', value: 'json'},
//		]}
//	>
//	<TabItem value="json">
//	```json
//	{
//	  "interface": {
//	    "method": {
//	      "annotations": [
//	        {
//	          "name": "http",
//	          "description": "Sets the method as an HTTP handler.",
//	          "parameters": [
//	            {
//	              "name": "method",
//	              "description": "Sets the HTTP method.",
//	              "type": "string",
//	              "required": true
//	              "validators": [
//	                {
//	                  "name": "HTTPMethodMustBeValid",
//	                  "expression": "{{includes (list `GET` `POST` `PUT` `DELETE` `PATCH` `HEAD` `OPTIONS` `TRACE` `CONNECT`) .}}",
//	                  "message": "http method must be valid"
//	                }
//	              ]
//	            }
//	          ]
//	        }
//	      ],
//	      "validators": [
//	        {
//	          "name": "MethodNameMustBeCapitalized",
//	          "expression": "{{eq .Name (.Name | capitalize)}}",
//	          "message": "method name must be capitalized"
//	        }
//	      ]
//	    }
//	  }
//	}
//	```
//	</TabItem>
//	<TabItem value="yaml">
//	```yaml
//	interface:
//	  method:
//	    annotations:
//	      - name: http
//	        description: Sets the method as an HTTP handler.
//	        parameters:
//	          - name: method
//	            description: Sets the HTTP method.
//	            types: [string]
//	            required: true
//	            validators:
//	              - name: HTTPMethodMustBeValid
//	                expression: "{{includes (list `GET` `POST` `PUT` `DELETE` `PATCH` `HEAD` `OPTIONS` `TRACE` `CONNECT`) .}}"
//	                message: http method must be valid
//	    validators:
//	      - name: MethodNameMustBeCapitalized
//	        expression: "{{eq .Name (.Name | capitalize)}}"
//	        message: method name must be capitalized
//	```
//	</TabItem>
//	</Tabs>
type InterfaceMethod struct {
	// @api(Grammar/InterfaceMethod.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the interface method declaration.
	Annotations Annotations `json:"annotations"`
	// @api(Grammar/InterfaceMethod.Validators) represents the [Validator](#Grammar/Common/Validator) for the interface method declaration.
	Validators Validators `json:"validators"`
	// @api(Grammar/InterfaceMethod.Parameter) represents the [InterfaceMethodParameter](#Grammar/InterfaceMethodParameter) grammar rules for the interface method declaration.
	Parameter InterfaceMethodParameter `json:"parameter"`
}

// @api(Grammar/InterfaceMethodParameter) represents the grammar rules for the interface method parameter declaration.
type InterfaceMethodParameter struct {
	// @api(Grammar/InterfaceMethodParameter.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the interface method parameter declaration.
	Annotations Annotations `json:"annotations"`
	// @api(Grammar/InterfaceMethodParameter.Validators) represents the [Validator](#Grammar/Common/Validator) for the interface method parameter declaration.
	Validators Validators `json:"validators"`
}

// Resolve resolves the grammar rules.
func (g *Grammar) Resolve() error {
	// Resolve package
	if err := g.doResolve("package", g.Package.Annotations, g.Package.Validators); err != nil {
		return err
	}
	// Resolve const
	for _, v := range g.Const.Types {
		if !slices.Contains(validConstTypes, v) {
			return fmt.Errorf("const.types: type %q must be one of %v", v, validConstTypes)
		}
	}
	if err := duplicated(g.Const.Types); err != nil {
		return fmt.Errorf("const.types: %w", err)
	}
	if err := g.doResolve("const", g.Const.Annotations, g.Const.Validators); err != nil {
		return err
	}
	// Resolve enum
	if err := g.doResolve("enum", g.Enum.Annotations, g.Enum.Validators); err != nil {
		return err
	}
	for _, v := range g.Enum.Member.Types {
		if !slices.Contains(validEnumMemberTypes, v) {
			return fmt.Errorf("enum.member.types: type %q must be one of %v", v, validEnumMemberTypes)
		}
	}
	if err := duplicated(g.Enum.Member.Types); err != nil {
		return fmt.Errorf("enum.member.types: %w", err)
	}
	if err := g.doResolve("enum.member", g.Enum.Member.Annotations, g.Enum.Member.Validators); err != nil {
		return err
	}
	// Resolve struct
	if err := g.doResolve("struct", g.Struct.Annotations, g.Struct.Validators); err != nil {
		return err
	}
	if err := g.doResolve("struct.field", g.Struct.Field.Annotations, g.Struct.Field.Validators); err != nil {
		return err
	}
	// Resolve interface
	if err := g.doResolve("interface", g.Interface.Annotations, g.Interface.Validators); err != nil {
		return err
	}
	if err := g.doResolve("interface.method", g.Interface.Method.Annotations, g.Interface.Method.Validators); err != nil {
		return err
	}
	if err := g.doResolve("interface.method.parameter", g.Interface.Method.Parameter.Annotations, g.Interface.Method.Parameter.Validators); err != nil {
		return err
	}

	if !g.builtin {
		g.extends(Builtin)
	}

	return nil
}

func (g *Grammar) extends(from Grammar) {
	appendTo(&g.Package.Annotations, from.Package.Annotations...)
	appendTo(&g.Package.Validators, from.Package.Validators...)
	appendTo(&g.Const.Annotations, from.Const.Annotations...)
	appendTo(&g.Const.Validators, from.Const.Validators...)
	appendTo(&g.Enum.Annotations, from.Enum.Annotations...)
	appendTo(&g.Enum.Validators, from.Enum.Validators...)
	appendTo(&g.Enum.Member.Annotations, from.Enum.Member.Annotations...)
	appendTo(&g.Enum.Member.Validators, from.Enum.Member.Validators...)
	appendTo(&g.Struct.Annotations, from.Struct.Annotations...)
	appendTo(&g.Struct.Validators, from.Struct.Validators...)
	appendTo(&g.Struct.Field.Annotations, from.Struct.Field.Annotations...)
	appendTo(&g.Struct.Field.Validators, from.Struct.Field.Validators...)
	appendTo(&g.Interface.Annotations, from.Interface.Annotations...)
	appendTo(&g.Interface.Validators, from.Interface.Validators...)
	appendTo(&g.Interface.Method.Annotations, from.Interface.Method.Annotations...)
	appendTo(&g.Interface.Method.Validators, from.Interface.Method.Validators...)
	appendTo(&g.Interface.Method.Parameter.Annotations, from.Interface.Method.Parameter.Annotations...)
	appendTo(&g.Interface.Method.Parameter.Validators, from.Interface.Method.Parameter.Validators...)
	if len(g.Const.Types) == 0 {
		g.Const.Types = from.Const.Types
	}
	if len(g.Enum.Member.Types) == 0 {
		g.Enum.Member.Types = from.Enum.Member.Types
	}
}

func (g *Grammar) doResolve(node string, annotations Annotations, validators Validators) error {
	if err := g.resolveAnnotations(node, annotations); err != nil {
		return err
	}
	if err := g.resolveValidators(node, validators); err != nil {
		return err
	}
	return nil
}

func (g *Grammar) resolveAnnotations(node string, annotations Annotations) error {
	for i := range annotations {
		if err := annotations[i].resolve("context.annotations", g.Context.Annotations); err != nil {
			return fmt.Errorf("%s.annotations: %w", node, err)
		}
		a := &annotations[i].value
		for j := range a.Parameters {
			if err := a.Parameters[j].resolve("context.annotation_parameters", g.Context.AnnotationParameters); err != nil {
				return fmt.Errorf("%s.annotations.parameters: %w", node, err)
			}
			p := &a.Parameters[j].value
			if err := g.resolveValidators(node, p.Validators); err != nil {
				return fmt.Errorf("%s: %w", node, err)
			}
		}
		if err := a.validate(g); err != nil {
			return fmt.Errorf("%s: %w", node, err)
		}
	}
	return nil
}

func (g *Grammar) resolveValidators(node string, validators Validators) error {
	for i := range validators {
		if err := validators[i].resolve("context.validators", g.Context.Validators); err != nil {
			return fmt.Errorf("%s: %w", node, err)
		}
	}
	return nil
}

func (a *Annotation) validate(g *Grammar) error {
	if a.Name == "" {
		return errors.New("annotation: name is required")
	}
	if !g.builtin && a.Name == "next" {
		return errors.New("annotation: \"next\" is reserved by the next compiler")
	}
	for i := range a.Parameters {
		p := &a.Parameters[i].value
		if err := p.validate(g); err != nil {
			return fmt.Errorf("annotation %s: %w", a.Name, err)
		}
	}
	return nil
}

func (p *AnnotationParameter) validate(_ *Grammar) error {
	if p.Name == "" {
		return fmt.Errorf("parameter: name is required")
	}
	if !stringutil.IsIdentifer(p.Name) {
		pattern, err := regexp.Compile("^" + p.Name + "$")
		if err != nil {
			return fmt.Errorf("parameter: invalid name pattern %q: %w", p.Name, err)
		}
		p.parsed.name = pattern
	}
	if len(p.Types) == 0 {
		return fmt.Errorf("parameter %s: types are empty", p.Name)
	}
	for _, t := range p.Types {
		if t == Any && len(p.Types) > 1 {
			return fmt.Errorf("parameter %s: only one type is allowed if any type is specified, got %d", p.Name, len(p.Types))
		}
		if !slices.Contains(validAnnotationParameterTypes, t) {
			return fmt.Errorf("parameter %s: type %s must be one of %v", p.Name, t, validAnnotationParameterTypes)
		}
	}
	return nil
}

const notEmpty = "{{ne . ``}}"

func next(node string, parameters ...Options[AnnotationParameter]) Options[Annotation] {
	return opt(Annotation{
		Name:        "next",
		Description: "Built-in annotation for " + node + " reserved by the next compiler.",
		Parameters:  parameters,
	})
}

func base_next(node string, parameters ...Options[AnnotationParameter]) Options[Annotation] {
	return next(node, append(parameters, available(), deprecated(), tokens())...)
}

func available() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "available",
		Description: "Sets the available expression for the declaration.",
		Types:       types(String),
	})
}

func deprecated() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "deprecated",
		Description: "Sets the declaration as deprecated.",
		Types:       types(String, Bool),
	})
}

func tokens() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "tokens",
		Description: "Sets the space separated tokens for the declaration.",
		Types:       types(String),
	})
}

func prompt() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "prompt",
		Description: "Sets the prompt message for the interface declaration.",
		Types:       types(String),
	})
}

func optional() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "optional",
		Description: "Sets the field as optional.",
		Types:       types(Bool),
	})
}

func default_() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "default",
		Description: "Sets the default value for field.",
		Types:       types(Any),
	})
}

func mut() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "mut",
		Description: "Sets object or parameter as mutable.",
		Types:       types(Bool),
	})
}

func error_() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "error",
		Description: "Indicates the method returns an error.",
		Types:       types(Bool),
	})
}

func type_() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "type",
		Description: "Sets the type for enum members.",
		Types:       types(Type),
	})
}

func lang_package() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        ".+_package",
		Description: "Sets the package name for target languages.",
		Types:       types(String),
		Validators: Validators{
			opt(Validator{
				Name:       "LangPackageMustNotBeEmpty",
				Expression: notEmpty,
				Message:    "must not be empty",
			}),
		},
	})
}

func lang_imports() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        ".+_imports",
		Description: "Sets the import declarations for target languages.",
		Types:       types(String),
	})
}

func lang_alias() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        ".+_alias",
		Description: "Sets the alias name for target languages.",
		Types:       types(String),
	})
}

func appendTo[S ~[]T, T any](s *S, x ...T) {
	*s = append(*s, x...)
}

// @api(Grammar/Builtin) represents the built-in grammar rules.
// Here is an example source code for demonstration built-in grammar rules:
//
//	```next
//	@next(
//		// available is used to set the available expression for the file.
//		available="cpp|java",
//		// deprecated is used to set the declaration as deprecated.
//		deprecated,
//		// tokens is used to set the space separated tokens for the declaration name.
//		tokens="DB",
//		// <LANG>_package is used to set the package name for target languages.
//		cpp_package="db",
//		java_package="com.example.db",
//		// <LANG>_imports is used to set the import declarations for target languages.
//		cpp_imports="<iostream>",
//	)
//	package db;
//
//	@next(
//		available="cpp",
//		deprecated,
//		tokens="HTTP Version",
//	)
//	const int HTTPVersion = 2;
//
//	@next(
//		available="cpp|java",
//		deprecated,
//		tokens="HTTP Method",
//		// <LANG>_alias is used to set the alias name for target languages.
//		java_alias="org.springframework.http.HttpMethod",
//	)
//	enum HTTPMethod {
//		Get = "GET",
//		Post = "POST",
//		Put = "PUT",
//		Delete = "DELETE",
//		Head = "HEAD",
//		Options = "OPTIONS",
//		Patch = "PATCH",
//		Trace = "TRACE",
//		@next(available="!java", deprecated, tokens="Connect")
//		Connect = "CONNECT",
//	}
//
//	@next(
//		available="cpp|java",
//		deprecated,
//		tokens="User",
//		cpp_alias="model::User",
//	)
//	struct User {
//		int id;
//		@next(
//			available="!java",
//			deprecated,
//			tokens="Name",
//			optional,
//			default="John Doe",
//			cpp_alias="std::string",
//		)
//		string name;
//	}
//
//	@next(
//		available="cpp|java",
//		deprecated,
//		tokens="User Repository",
//		prompt="Prompt for AGI.",
//		java_alias="org.springframework.data.repository.CrudRepository<User, Long>",
//	)
//	interface UserRepository {
//		@next(
//			available="!java",
//			deprecated,
//			tokens="Get User",
//			mut,
//			error,
//			cpp_alias="std::shared_ptr<model::User>",
//			prompt="Prompt for AGI.",
//		)
//		findById(
//			@next(available="!java", deprecated, tokens="ID", mut, cpp_alias="int64_t")
//			int64 id
//		) User;
//	}
//	```
var Builtin = Grammar{
	builtin: true,
	Package: Package{
		Annotations: Annotations{base_next("package", lang_package(), lang_imports())},
	},
	Const: Const{
		Annotations: Annotations{base_next("const")},
		Types:       validConstTypes,
	},
	Enum: Enum{
		Annotations: Annotations{base_next("enum", type_())},
		Member: EnumMember{
			Annotations: Annotations{base_next("enum.member")},
			Types:       validEnumMemberTypes,
		},
	},
	Struct: Struct{
		Annotations: Annotations{base_next("struct", lang_alias())},
		Field: StructField{
			Annotations: Annotations{base_next("struct.field", optional(), default_(), lang_alias())},
		},
	},
	Interface: Interface{
		Annotations: Annotations{base_next("interface", prompt(), lang_alias())},
		Method: InterfaceMethod{
			Annotations: Annotations{base_next("interface.method", prompt(), lang_alias(), mut(), error_())},
			Parameter: InterfaceMethodParameter{
				Annotations: Annotations{next("interface.method.parameter", deprecated(), tokens(), mut(), lang_alias())},
			},
		},
	},
}

func init() {
	if err := Builtin.Resolve(); err != nil {
		panic(err)
	}
}

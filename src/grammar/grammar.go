//go:generate go run github.com/gopherd/tools/cmd/docgen@v0.0.8 -ext .mdx -I ./ -o ../../website/docs/api/preview -level 0 -M "---" -M "pagination_prev: null" -M "pagination_next: null" -M "---"
package grammar

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"github.com/gopherd/core/container/iters"
	"github.com/gopherd/core/text/templates"
	"github.com/next/next/src/internal/stringutil"
)

const (
	Bool   = "bool"
	Int    = "int"
	Float  = "float"
	String = "string"
	Type   = "type"
)

var validateAnnotationParameterTypes = []string{Bool, Int, Float, String, Type}
var validateConstTypes = []string{Bool, Int, Float, String}
var validateEnumMemberTypes = []string{Int, Float, String}

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

func at(ids ...string) Annotations {
	return slices.Collect(iters.Map(slices.Values(ids), func(id string) Options[Annotation] {
		return Options[Annotation]{id: id}
	}))
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

func (o *Options[T]) resolve(source map[string]json.RawMessage) error {
	if o.id == "" {
		return nil
	}
	v, ok := source[o.id]
	if !ok {
		return fmt.Errorf("options %q not found", o.id)
	}
	return json.Unmarshal(v, &o.value)
}

// @api(Grammar) represents the custom grammar for the next files.
//
//	import CodeBlock from "@theme/CodeBlock";
//	import Tabs from "@theme/Tabs";
//	import TabItem from "@theme/TabItem";
//	import ExampleGrammarYAMLSource from "!!raw-loader!@site/example/grammar.yaml";
//
// The grammar is used to define a subset of the next files. It can limit the features of the next code according
// by your requirements. The grammar is a yaml file that contains rules.
//
// Here is an example of the grammar file:
//
//	<CodeBlock language="yaml" title="grammar.yaml">
//		{ExampleGrammarYAMLSource}
//	</CodeBlock>
type Grammar struct {
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
//	          type: int
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

	// @api(Grammar/Common/AnnotationParameter.Type) represents the parameter type.
	// The type is a string that can be one of the following types:
	//
	// - **bool**: boolean type, the value can be `true` or `false`
	// - **int**: integer type, the value can be a positive or negative integer, for example, `123`
	// - **float**: float type, the value can be a positive or negative float, for example, `1.23`
	// - **string**: string type, the value can be a string, for example, `"hello"`
	// - **type**: any type name, for example, `int`, `float`, `string`, etc. Custom type names are supported.
	Type string `json:"type"`

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
//	            type: string
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
		if !slices.Contains(validateConstTypes, v) {
			return fmt.Errorf("const type %q must be one of %v", v, validateConstTypes)
		}
	}
	if err := duplicated(g.Const.Types); err != nil {
		return fmt.Errorf("const types: %w", err)
	}
	if err := g.doResolve("const", g.Const.Annotations, g.Const.Validators); err != nil {
		return err
	}
	// Resolve enum
	if err := g.doResolve("enum", g.Enum.Annotations, g.Enum.Validators); err != nil {
		return err
	}
	for _, v := range g.Enum.Member.Types {
		if !slices.Contains(validateEnumMemberTypes, v) {
			return fmt.Errorf("enum member type %q must be one of %v", v, validateEnumMemberTypes)
		}
	}
	if err := duplicated(g.Enum.Member.Types); err != nil {
		return fmt.Errorf("enum member types: %w", err)
	}
	if err := g.doResolve("enum member", g.Enum.Member.Annotations, g.Enum.Member.Validators); err != nil {
		return err
	}
	// Resolve struct
	if err := g.doResolve("struct", g.Struct.Annotations, g.Struct.Validators); err != nil {
		return err
	}
	if err := g.doResolve("struct field", g.Struct.Field.Annotations, g.Struct.Field.Validators); err != nil {
		return err
	}
	// Resolve interface
	if err := g.doResolve("interface", g.Interface.Annotations, g.Interface.Validators); err != nil {
		return err
	}
	if err := g.doResolve("interface method", g.Interface.Method.Annotations, g.Interface.Method.Validators); err != nil {
		return err
	}
	if err := g.doResolve("interface method parameter", g.Interface.Method.Parameter.Annotations, g.Interface.Method.Parameter.Validators); err != nil {
		return err
	}
	return nil
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
		if err := annotations[i].resolve(g.Context.Annotations); err != nil {
			return fmt.Errorf("%s: %w", node, err)
		}
		a := &annotations[i].value
		for j := range a.Parameters {
			if err := a.Parameters[j].resolve(g.Context.AnnotationParameters); err != nil {
				return fmt.Errorf("%s: %w", node, err)
			}
			p := &a.Parameters[j].value
			if err := g.resolveValidators(node, p.Validators); err != nil {
				return fmt.Errorf("%s: %w", node, err)
			}
		}
		if err := a.validate(); err != nil {
			return fmt.Errorf("%s: %w", node, err)
		}
	}
	return nil
}

func (g *Grammar) resolveValidators(node string, validators Validators) error {
	for i := range validators {
		if err := validators[i].resolve(g.Context.Validators); err != nil {
			return fmt.Errorf("%s: %w", node, err)
		}
	}
	return nil
}

func (a *Annotation) validate() error {
	if a.Name == "" {
		return fmt.Errorf("annotation name is required")
	}
	for i := range a.Parameters {
		p := &a.Parameters[i].value
		if err := p.validate(); err != nil {
			return fmt.Errorf("annotation %q: %w", a.Name, err)
		}
	}
	return nil
}

func (p *AnnotationParameter) validate() error {
	if p.Name == "" {
		return fmt.Errorf("parameter name is required")
	}
	if !stringutil.IsIdentifer(p.Name) {
		pattern, err := regexp.Compile("^" + p.Name + "$")
		if err != nil {
			return fmt.Errorf("invalid parameter name pattern %q: %w", p.Name, err)
		}
		p.parsed.name = pattern
	}
	if p.Type == "" {
		return fmt.Errorf("parameter %q: type is required", p.Name)
	}
	if slices.Contains(validateAnnotationParameterTypes, p.Type) {
		return nil
	}
	return fmt.Errorf("type %q must be one of %v", p.Type, validateAnnotationParameterTypes)
}

const notEmpty = "{{ne . ``}}"

func toJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func next(parameters ...string) Options[Annotation] {
	return opt(Annotation{
		Name:        "next",
		Description: "Built-in annotation for the next compiler.",
		Parameters: slices.Collect(iters.Map(slices.Values(parameters), func(id string) Options[AnnotationParameter] {
			return Options[AnnotationParameter]{id: id}
		})),
	})
}

func deprecated() Annotation {
	return Annotation{
		Name:        "deprecated",
		Description: "Sets the declaration as deprecated.",
		Parameters: []Options[AnnotationParameter]{
			opt(AnnotationParameter{
				Name:        "message",
				Description: "Sets the deprecation message.",
				Type:        String,
			}),
		},
	}
}

func required() Annotation {
	return Annotation{
		Name:        "required",
		Description: "Sets the field as required.",
	}
}

func optional() Annotation {
	return Annotation{
		Name:        "optional",
		Description: "Sets the field as optional.",
		Parameters:  []Options[AnnotationParameter]{default_()},
	}
}

func default_() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "default",
		Description: "Sets the default value for optional fields.",
		Type:        String,
	})
}

func snake_case() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "snake_case",
		Description: "Sets the snake_case name for the declaration.",
		Type:        String,
	})
}

func camel_case() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "camel_case",
		Description: "Sets the camelCase name for the declaration.",
		Type:        String,
	})
}

func pascal_case() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "pascal_case",
		Description: "Sets the PascalCase name for the declaration.",
		Type:        String,
	})
}

func kebab_case() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "kebab_case",
		Description: "Sets the kebab-case name for the declaration.",
		Type:        String,
	})
}

func available() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "available",
		Description: "Sets the available expression for the declaration.",
		Type:        String,
	})
}

func mut() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "mut",
		Description: "Sets object or parameter as mutable.",
		Type:        Bool,
	})
}

func error_() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "error",
		Description: "Indicates the method returns an error.",
		Type:        Bool,
	})
}

func type_() Options[AnnotationParameter] {
	return opt(AnnotationParameter{
		Name:        "type",
		Description: "Sets the type for enum members.",
		Type:        Type,
	})
}

func LANG_package() AnnotationParameter {
	return AnnotationParameter{
		Name:        ".+_package",
		Description: "Sets the package name for target languages.",
		Type:        String,
		Validators: Validators{
			opt(Validator{
				Name:       "LangPackageMustNotBeEmpty",
				Expression: notEmpty,
				Message:    "must not be empty",
			}),
		},
	}
}

func LANG_imports() AnnotationParameter {
	return AnnotationParameter{
		Name:        ".+_imports",
		Description: "Sets the import declarations for target languages.",
		Type:        String,
	}
}

func LANG_alias() AnnotationParameter {
	return AnnotationParameter{
		Name:        ".+_alias",
		Description: "Sets the alias name for target languages.",
		Type:        String,
	}
}

func LANG_type() AnnotationParameter {
	return AnnotationParameter{
		Name:        ".+_type",
		Description: "Sets the type name for target languages.",
		Type:        String,
	}
}

// @api(Grammar.default) represents the default grammar for the next files.
//
//	:::tip
//
//	Run the following command to generate the default grammar:
//
//	```sh
//	next grammar grammar.yaml
//	```
//
//	You can use this grammar as a starting point for your own grammar.
//
//	:::
var Default = Grammar{
	Context: Context{
		Annotations: map[string]json.RawMessage{
			"next@package":                    toJSON(next("available", "*_package", "*_imports", "snake_case", "camel_case", "pascal_case", "kebab_case")),
			"next@const":                      toJSON(next("available", "snake_case", "camel_case", "pascal_case", "kebab_case")),
			"next@enum":                       toJSON(next("available", "type")),
			"next@enum.member":                toJSON(next("available", "snake_case", "camel_case", "pascal_case", "kebab_case")),
			"next@struct":                     toJSON(next("available", "*_alias")),
			"next@struct.field":               toJSON(next("available", "*_type", "snake_case", "camel_case", "pascal_case", "kebab_case")),
			"next@interface":                  toJSON(next("available", "*_alias")),
			"next@interface.method":           toJSON(next("available", "mut", "error", "snake_case", "camel_case", "pascal_case", "kebab_case")),
			"next@interface.method.parameter": toJSON(next("mut", "*_type", "snake_case", "camel_case", "pascal_case", "kebab_case")),
			"deprecated":                      toJSON(deprecated()),
			"required":                        toJSON(required()),
			"optional":                        toJSON(optional()),
		},
		AnnotationParameters: map[string]json.RawMessage{
			"available":   toJSON(available()),
			"mut":         toJSON(mut()),
			"error":       toJSON(error_()),
			"type":        toJSON(type_()),
			"snake_case":  toJSON(snake_case()),
			"camel_case":  toJSON(camel_case()),
			"pascal_case": toJSON(pascal_case()),
			"kebab_case":  toJSON(kebab_case()),
			"*_package":   toJSON(LANG_package()),
			"*_imports":   toJSON(LANG_imports()),
			"*_alias":     toJSON(LANG_alias()),
			"*_type":      toJSON(LANG_type()),
		},
	},
	Package: Package{
		Annotations: at("next@package", "deprecated"),
	},
	Const: Const{
		Types:       []string{Bool, Int, Float, String},
		Annotations: at("next@const", "deprecated"),
	},
	Enum: Enum{
		Annotations: at("next@enum", "deprecated"),
		Member: EnumMember{
			Types:       []string{Int, Float, String},
			Annotations: at("next@enum.member", "deprecated"),
		},
	},
	Struct: Struct{
		Annotations: at("next@struct", "deprecated"),
		Field: StructField{
			Annotations: at("next@struct.field", "deprecated", "required", "optional"),
		},
	},
	Interface: Interface{
		Annotations: at("next@interface", "deprecated"),
		Method: InterfaceMethod{
			Annotations: at("next@interface.method", "deprecated"),
			Parameter: InterfaceMethodParameter{
				Annotations: at("next@interface.method.parameter", "deprecated"),
			},
		},
	},
}

func init() {
	// Resolve the default grammar
	if err := Default.Resolve(); err != nil {
		panic(err)
	}
}

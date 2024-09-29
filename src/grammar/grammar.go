//go:generate go run github.com/gopherd/tools/cmd/docgen@v0.0.8 -ext .mdx -I ./ -o ../../website/docs/api/preview -level 0 -M "---" -M "pagination_prev: null" -M "pagination_next: null" -M "---"
package grammar

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"github.com/gopherd/core/text/templates"
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

// @api(Grammar) represents the custom grammar for the next files.
//
// The grammar is used to define a subset of the next files. It can limit the features of the next code according
// by your requirements. The grammar is a JSON object that contains rules.
type Grammar struct {
	Package   Package
	Import    Import
	Const     Const
	Enum      Enum
	Struct    Struct
	Interface Interface
}

// @api(Grammar/Common/Validator) represents the validator for the grammar rules.
type Validator struct {
	// @api(Grammar/Common/Validator.Name) represents the validator name.
	Name string

	// @api(Grammar/Common/Validator.Expression) represents the validator expression.
	// The expression is a template string that can access the data by the `.` operator. The expression
	// must return a boolean value.
	//
	// The data is the current context object. For example, **package** object for the package validator.
	Expression string

	// @api(Grammar/Common/Validator.Message) represents the error message when the validator is failed.
	Message string
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
//	```json
//	{
//	  "Struct": {
//	    "Annotations": [
//	      {
//	        "Name": "message",
//	        "Description": "Sets the struct as a message.",
//	        "Parameters": [
//	          {
//	            "Name": "type",
//	            "Description": "Sets the message type id.",
//	            "Type": "int",
//	            "Required": true
//	            "Validators": [
//	              {
//	                "Name": "MessageTypeMustBePositive",
//	                "Expression": "{{gt . 0}}",
//	                "Message": "message type must be positive"
//	              }
//	            ]
//	          }
//	        ]
//	      }
//	    ]
//	  }
//	}
//	```
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
	Name string

	// @api(Grammar/Common/Annotation.Description) represents the annotation description.
	Description string

	// @api(Grammar/Common/Annotation.Parameters) represents the annotation parameters.
	Parameters []AnnotationParameter `json:",omitempty"`
}

func LookupAnnotation(annotations []Annotation, name string) *Annotation {
	for i := range annotations {
		a := &annotations[i]
		if a.Name == name {
			return a
		}
	}
	return nil
}

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
	Name string

	// @api(Grammar/Common/AnnotationParameter.Description) represents the parameter description.
	Description string

	// @api(Grammar/Common/AnnotationParameter.Type) represents the parameter type.
	// The type is a string that can be one of the following types:
	//
	// - **bool**: boolean type, the value can be `true` or `false`
	// - **int**: integer type, the value can be a positive or negative integer, for example, `123`
	// - **float**: float type, the value can be a positive or negative float, for example, `1.23`
	// - **string**: string type, the value can be a string, for example, `"hello"`
	// - **type**: any type name, for example, `int`, `float`, `string`, etc. Custom type names are supported.
	Type string

	// @api(Grammar/Common/AnnotationParameter.Required) represents the parameter is required or not.
	Required bool `json:",omitempty"`

	// @api(Grammar/Common/AnnotationParameter.Validators) represents the [Validator](#Grammar/Common/Validator) for the annotation parameter.
	Validators []Validator `json:",omitempty"`

	parsed struct {
		name *regexp.Regexp
	} `json:"-"`
}

func isIdentifer(s string) bool {
	r := []rune(s)
	if len(r) == 0 {
		return false
	}
	if !isLetter(r[0]) || r[0] == '_' {
		return false
	}
	for i := 1; i < len(r); i++ {
		if !isLetter(r[i]) && !isDigit(r[i]) && r[i] != '_' {
			return false
		}
	}
	return true
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z'
}

func LookupAnnotationParameter(parameters []AnnotationParameter, name string) *AnnotationParameter {
	for i := range parameters {
		p := &parameters[i]
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
	Annotations []Annotation `json:",omitempty"`

	// @api(Grammar/Package.Validators) represents the [Validator](#Grammar/Common/Validator) for the package declaration.
	// It's used to validate the package name. For example, You can limit the package name must be
	// not start with a "_" character. The validator expression is a template string that can access
	// package name by `.Name`.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Package": {
	//	    "Validators": [
	//	      {
	//	        "Name": "PackageNameNotStartWithUnderscore",
	//	        "Expression": "{{not (hasPrefix `_` .Name)}}",
	//	        "Message": "package name must not start with an underscore"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
	//
	//	```next
	//	// This will error
	//	package _test;
	//	// Error: package name must not start with an underscore
	//	```
	Validators []Validator `json:",omitempty"`
}

// @api(Grammar/Import) represents the grammar rules for the import declaration.
type Import struct {
	// @api(Grammar/Import.Off) represents the import declaration is off or not.
	// If the import declaration is off, the import declaration is not allowed in the next files.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Import": {
	//	    "Off": true
	//	  }
	//	}
	//	```
	//
	//	```next
	//	package demo;
	//
	//	// This will error
	//	import "other.next";
	//	// Error: import declaration is not allowed
	//	```
	Off bool `json:",omitempty"`
}

// @api(Grammar/Const) represents the grammar rules for the const declaration.
type Const struct {
	// @api(Grammar/Const.Off) represents the const declaration is off or not.
	// If the const declaration is off, the const declaration is not allowed in the next files.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Const": {
	//	    "Off": true
	//	  }
	//	}
	//	```
	//
	//	```next
	//	package demo;
	//
	//	// This will error
	//	const x = 1;
	//	// Error: const declaration is not allowed
	//	```
	Off bool `json:",omitempty"`

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
	//	```json
	//	{
	//	  "Const": {
	//	    "Types": ["int", "float"]
	//	  }
	//	}
	//	```
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
	Types []string `json:",omitempty"`

	// @api(Grammar/Const.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the const declaration.
	Annotations []Annotation `json:",omitempty"`

	// @api(Grammar/Const.Validators) represents the [Validator](#Grammar/Common/Validator) for the const declaration.
	// It's used to validate the const name. You can access the const name by `.Name`.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Const": {
	//	    "Validators": [
	//	      {
	//	        "Name": "ConstNameMustBeCapitalized",
	//	        "Expression": "{{eq .Name (.Name | capitalize)}}",
	//	        "Message": "const name must be capitalized"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
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
	Validators []Validator `json:",omitempty"`
}

// @api(Grammar/Enum) represents the grammar rules for the enum declaration.
type Enum struct {
	// @api(Grammar/Enum.Off) represents the enum declaration is off or not.
	// If the enum declaration is off, the enum declaration is not allowed in the next files.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Enum": {
	//	    "Off": true
	//	  }
	//	}
	//	```
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
	Off bool `json:",omitempty"`

	// @api(Grammar/Enum.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the enum declaration.
	Annotations []Annotation `json:",omitempty"`

	// @api(Grammar/Enum.Validators) represents the [Validator](#Grammar/Common/Validator) for the enum declaration.
	// It's used to validate the enum name. You can access the enum name by `.Name`.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Enum": {
	//	    "Validators": [
	//	      {
	//	        "Name": "EnumNameMustBeCapitalized",
	//	        "Expression": "{{eq .Name (.Name | capitalize)}}",
	//	        "Message": "enum name must be capitalized"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
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
	Validators []Validator `json:",omitempty"`

	// @api(Grammar/Enum.Member) represents the [EnumMember](#Grammar/EnumMember) grammar rules for the enum declaration.
	Member EnumMember
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
	//	```json
	//	{
	//	  "Enum": {
	//	    "Member": {
	//	      "Types": ["int"]
	//	    }
	//	  }
	//	}
	//	```
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
	Types []string `json:",omitempty"`

	// @api(Grammar/EnumMember.ValueRequired) represents the enum member value is required or not.
	// If the enum member value is required, the enum member value must be specified in the next files.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Enum": {
	//	    "Member": {
	//	      "ValueRequired": true
	//	    }
	//	  }
	//	}
	//	```
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
	ValueRequired bool `json:",omitempty"`

	// @api(Grammar/EnumMember.ZeroRequired) represents the enum member zero value for integer types is required or not.
	//
	// If the enum member zero value is required, the enum member zero value must be specified in the next files.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Enum": {
	//	    "Member": {
	//	      "ZeroRequired": true
	//	    }
	//	  }
	//	}
	//	```
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
	ZeroRequired bool `json:",omitempty"`

	// @api(Grammar/EnumMember.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the enum member declaration.
	Annotations []Annotation `json:",omitempty"`

	// @api(Grammar/EnumMember.Validators) represents the [Validator](#Grammar/Common/Validator) for the enum member declaration.
	//
	// It's used to validate the enum member name. You can access the enum member name by `.Name`.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Enum": {
	//	    "Member": {
	//	      "Validators": [
	//	        {
	//	          "Name": "EnumMemberNameMustBeCapitalized",
	//	          "Expression": "{{eq .Name (.Name | capitalize)}}",
	//	          "Message": "enum member name must be capitalized"
	//	        }
	//	      ]
	//	    }
	//	  }
	//	}
	//	```
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
	Validators []Validator `json:",omitempty"`
}

// @api(Grammar/Struct) represents the grammar rules for the struct declaration.
type Struct struct {
	// @api(Grammar/Struct.Off) represents the struct declaration is off or not.
	// If the struct declaration is off, the struct declaration is not allowed in the next files.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Struct": {
	//	    "Off": true
	//	  }
	//	}
	//	```
	//
	//	```next
	//	package demo;
	//
	//	// This will error
	//	struct User {
	//	  ID int;
	//	  Name string;
	//	}
	//	// Error: struct declaration is not allowed
	//	```
	Off bool `json:",omitempty"`

	// @api(Grammar/Struct.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the struct declaration.
	Annotations []Annotation `json:",omitempty"`

	// @api(Grammar/Struct.Validators) represents the [Validator](#Grammar/Common/Validator) for the struct declaration.
	// It's used to validate the struct name. You can access the struct name by `.Name`.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Struct": {
	//	    "Validators": [
	//	      {
	//	        "Name": "StructNameMustBeCapitalized",
	//	        "Expression": "{{eq .Name (.Name | capitalize)}}",
	//	        "Message": "struct name must be capitalized"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
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
	Validators []Validator `json:",omitempty"`

	// @api(Grammar/Struct.Field) represents the [StructField](#Grammar/StructField) grammar rules for the struct declaration.
	Field StructField
}

// @api(Grammar/StructField) represents the grammar rules for the struct field declaration.
type StructField struct {
	// @api(Grammar/StructField.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the struct field declaration.
	Annotations []Annotation `json:",omitempty"`

	// @api(Grammar/StructField.Validators) represents the [Validator](#Grammar/Common/Validator) for the struct field declaration.
	// It's used to validate the struct field name. You can access the struct field name by `.Name`.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Struct": {
	//	    "Field": {
	//	      "Validators": [
	//	        {
	//	          "Name": "StructFieldNameMustNotBeCapitalized",
	//	          "Expression": "{{ne .Name (capitalize .Name)}}",
	//	          "Message": "struct field name must not be capitalized"
	//	        }
	//	      ]
	//	    }
	//	  }
	//	}
	//	```
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
	Validators []Validator `json:",omitempty"`
}

// @api(Grammar/Interface) represents the grammar rules for the interface declaration.
type Interface struct {
	// @api(Grammar/Interface.Off) represents the interface declaration is off or not.
	// If the interface declaration is off, the interface declaration is not allowed in the next files.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Interface": {
	//	    "Off": true
	//	  }
	//	}
	//	```
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
	Off bool `json:",omitempty"`

	// @api(Grammar/Interface.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the interface declaration.
	Annotations []Annotation `json:",omitempty"`

	// @api(Grammar/Interface.Validators) represents the [Validator](#Grammar/Common/Validator) for the interface declaration.
	// It's used to validate the interface name. You can access the interface name by `.Name`.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Interface": {
	//	    "Validators": [
	//	      {
	//	        "Name": "InterfaceNameMustBeCapitalized",
	//	        "Expression": "{{eq .Name (.Name | capitalize)}}",
	//	        "Message": "interface name must be capitalized"
	//	      }
	//	    ]
	//	  }
	//	}
	//	```
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
	Validators []Validator `json:",omitempty"`

	// @api(Grammar/Interface.Method) represents the [InterfaceMethod](#Grammar/InterfaceMethod) grammar rules for the interface declaration.
	Method InterfaceMethod
}

// @api(Grammar/InterfaceMethod) represents the grammar rules for the interface method declaration.
//
// Example:
//
//	```json
//	{
//	  "Interface": {
//	    "Method": {
//	      "Annotations": [
//	        {
//	          "Name": "http",
//	          "Description": "Sets the method as an HTTP handler.",
//	          "Parameters": [
//	            {
//	              "Name": "method",
//	              "Description": "Sets the HTTP method.",
//	              "Type": "string",
//	              "Required": true
//	              "Validators": [
//	                {
//	                  "Name": "HTTPMethodMustBeValid",
//	                  "Expression": "{{includes (list `GET` `POST` `PUT` `DELETE` `PATCH` `HEAD` `OPTIONS` `TRACE` `CONNECT`) .}}",
//	                  "Message": "http method must be valid"
//	                }
//	              ]
//	            }
//	          ]
//	        }
//	      ],
//	      "Validators": [
//	        {
//	          "Name": "MethodNameMustBeCapitalized",
//	          "Expression": "{{eq .Name (.Name | capitalize)}}",
//	          "Message": "method name must be capitalized"
//	        }
//	      ]
//	    }
//	  }
//	}
//	```
type InterfaceMethod struct {
	// @api(Grammar/InterfaceMethod.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the interface method declaration.
	Annotations []Annotation `json:",omitempty"`
	// @api(Grammar/InterfaceMethod.Validators) represents the [Validator](#Grammar/Common/Validator) for the interface method declaration.
	Validators []Validator `json:",omitempty"`

	// @api(Grammar/InterfaceMethod/Parameter) represents the grammar rules for the interface method parameter declaration.
	Parameter struct {
		// @api(Grammar/InterfaceMethod/Parameter.Annotations) represents the [Annotation](#Grammar/Common/Annotation) grammar rules for the interface method parameter declaration.
		Annotations []Annotation `json:",omitempty"`
		// @api(Grammar/InterfaceMethod/Parameter.Validators) represents the [Validator](#Grammar/Common/Validator) for the interface method parameter declaration.
		Validators []Validator `json:",omitempty"`
	}
}

// Validate validates the grammar rules.
func (g *Grammar) Validate() error {
	// Validate package
	if err := validateAnnotations("package", g.Package.Annotations); err != nil {
		return err
	}
	// Validate const
	for _, v := range g.Const.Types {
		if !slices.Contains(validateConstTypes, v) {
			return fmt.Errorf("const type %q must be one of %v", v, validateConstTypes)
		}
	}
	if err := duplicated(g.Const.Types); err != nil {
		return fmt.Errorf("const types: %w", err)
	}
	if err := validateAnnotations("const", g.Const.Annotations); err != nil {
		return err
	}
	// Validate enum
	if err := validateAnnotations("enum", g.Enum.Annotations); err != nil {
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
	if err := validateAnnotations("enum member", g.Enum.Member.Annotations); err != nil {
		return err
	}
	// Validate struct
	if err := validateAnnotations("struct", g.Struct.Annotations); err != nil {
		return err
	}
	if err := validateAnnotations("struct field", g.Struct.Field.Annotations); err != nil {
		return err
	}
	// Validate interface
	if err := validateAnnotations("interface", g.Interface.Annotations); err != nil {
		return err
	}
	if err := validateAnnotations("interface method", g.Interface.Method.Annotations); err != nil {
		return err
	}
	if err := validateAnnotations("interface method parameter", g.Interface.Method.Parameter.Annotations); err != nil {
		return err
	}
	return nil
}

func validateAnnotations(node string, annotations []Annotation) error {
	for i := range annotations {
		a := &annotations[i]
		if err := a.validate(); err != nil {
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
		p := &a.Parameters[i]
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
	if !isIdentifer(p.Name) {
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

const notEmpty = "{{ne .Name ``}}"

func next(parameters ...AnnotationParameter) Annotation {
	return Annotation{
		Name:        "next",
		Description: "Built-in annotation for the next compiler.",
		Parameters:  parameters,
	}
}

func deprecated() Annotation {
	return Annotation{
		Name:        "deprecated",
		Description: "Sets the declaration as deprecated.",
		Parameters: []AnnotationParameter{
			{
				Name:        "message",
				Description: "Sets the deprecation message.",
				Type:        String,
			},
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
		Parameters:  []AnnotationParameter{default_()},
	}
}

func default_() AnnotationParameter {
	return AnnotationParameter{
		Name:        "default",
		Description: "Sets the default value for optional fields.",
		Type:        String,
	}
}

func available() AnnotationParameter {
	return AnnotationParameter{
		Name:        "available",
		Description: "Sets the available expression for the declaration.",
		Type:        String,
	}
}

func mut() AnnotationParameter {
	return AnnotationParameter{
		Name:        "mut",
		Description: "Sets object or parameter as mutable.",
		Type:        Bool,
	}
}

func error_() AnnotationParameter {
	return AnnotationParameter{
		Name:        "error",
		Description: "Indicates the method returns an error.",
		Type:        Bool,
	}
}

func type_() AnnotationParameter {
	return AnnotationParameter{
		Name:        "type",
		Description: "Sets the type for enum members.",
		Type:        Type,
	}
}

func LANG_package() AnnotationParameter {
	return AnnotationParameter{
		Name:        ".+_package",
		Description: "Sets the package name for target languages.",
		Type:        String,
		Validators: []Validator{
			{
				Name:       "LangPackageMustNotBeEmpty",
				Expression: notEmpty,
				Message:    "must not be empty",
			},
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
		Description: "Sets the field type name for target languages.",
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
//	next grammar grammar.json
//	```
//
//	You can use this grammar as a starting point for your own grammar.
//
//	:::
var Default = Grammar{
	Package: Package{
		Annotations: []Annotation{
			next(available(), LANG_package(), LANG_imports()),
			deprecated(),
		},
	},
	Const: Const{
		Types: []string{Bool, Int, Float, String},
		Annotations: []Annotation{
			next(available()),
			deprecated(),
		},
	},
	Enum: Enum{
		Annotations: []Annotation{
			next(available(), type_()),
			deprecated(),
		},
		Member: EnumMember{
			Types: []string{Int, Float, String},
			Annotations: []Annotation{
				next(available()),
				deprecated(),
			},
		},
	},
	Struct: Struct{
		Annotations: []Annotation{
			next(available(), LANG_alias()),
			deprecated(),
		},
		Field: StructField{
			Annotations: []Annotation{
				next(available(), LANG_type()),
				deprecated(),
				required(),
				optional(),
			},
		},
	},
	Interface: Interface{
		Annotations: []Annotation{
			next(available(), LANG_alias()),
			deprecated(),
		},
		Method: InterfaceMethod{
			Annotations: []Annotation{
				next(available(), mut(), error_()),
				deprecated(),
			},
			Parameter: struct {
				Annotations []Annotation `json:",omitempty"`
				Validators  []Validator  `json:",omitempty"`
			}{
				Annotations: []Annotation{
					next(available(), mut()),
					deprecated(),
				},
			},
		},
	},
}

func init() {
	// Ensure the default grammar is valid
	if err := Default.Validate(); err != nil {
		panic(err)
	}
}

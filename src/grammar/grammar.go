//go:generate go run github.com/gopherd/tools/cmd/docgen@v0.0.8 -ext .mdx -I ./ -o ../../website/docs/api/preview -level 0 -M "---" -M "pagination_prev: null" -M "pagination_next: null" -M "---"
package grammar

const (
	Bool   = "bool"
	Int    = "int"
	Float  = "float"
	String = "string"
	Type   = "type"
)

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
	// @api(Grammar/Common/Annotation.Name) represents the annotation name pattern.
	Name string

	// @api(Grammar/Common/Annotation.Description) represents the annotation description.
	Description string

	// @api(Grammar/Common/Annotation.Parameters) represents the annotation parameters.
	Parameters []AnnotationParameter `json:",omitempty"`
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
	// - "**\*_package**": matches the annotation name that ends with `_package`, for example, `cpp_package`, `java_package`, etc.
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
	// @api(Grammar/EnumMember.OffValueExpr) represents the enum member value expression is off or not.
	// If the enum member value expression is off, the enum member value expression is not allowed in the next files.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Enum": {
	//	    "Member": {
	//	      "OffValueExpr": true
	//	    }
	//	  }
	//	}
	//	```
	//
	//	```next
	//	package demo;
	//
	//	enum Size {
	//	  Small;
	//	  Medium;
	//	  // This will error
	//	  Large = Small + Medium;
	//	  // Error: enum member value expression is not allowed
	//	}
	//	```
	OffValueExpr bool `json:",omitempty"`

	// @api(Grammar/EnumMember.OffIota) represents the enum member iota value is off or not.
	// If the enum member iota value is off, the enum member iota value is not allowed in the next files.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Enum": {
	//	    "Member": {
	//	      "OffIota": true
	//	    }
	//	  }
	//	}
	//	```
	//
	//	```next
	//	package demo;
	//
	//	enum Size {
	//	  // This will error
	//	  Small = iota;
	//	  // Error: enum member iota value is not allowed
	//	  Medium;
	//	  Large;
	//	}
	//	```
	OffIota bool `json:",omitempty"`

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
	}
}

func available() AnnotationParameter {
	return AnnotationParameter{
		Name:        "available",
		Description: "Sets the available target languages.",
		Type:        Bool,
	}
}

func mut() AnnotationParameter {
	return AnnotationParameter{
		Name:        "mut",
		Description: "Sets object or parameter as mutable.",
		Type:        Bool,
	}
}

func error() AnnotationParameter {
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

func star_package() AnnotationParameter {
	return AnnotationParameter{
		Name:        "*_package",
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

func star_imports() AnnotationParameter {
	return AnnotationParameter{
		Name:        "*_imports",
		Description: "Sets the import declarations for target languages.",
		Type:        String,
	}
}

func star_alias() AnnotationParameter {
	return AnnotationParameter{
		Name:        "*_alias",
		Description: "Sets the alias name for target languages.",
		Type:        String,
	}
}

// Default represents the default grammar for the next files.
var Default = Grammar{
	Package: Package{
		Annotations: []Annotation{
			next(available(), star_package(), star_imports()),
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
			next(available(), star_alias()),
			deprecated(),
		},
		Field: StructField{
			Annotations: []Annotation{
				next(available()),
				deprecated(),
				required(),
				optional(),
			},
		},
	},
	Interface: Interface{
		Annotations: []Annotation{
			next(available(), star_alias()),
			deprecated(),
		},
		Method: InterfaceMethod{
			Annotations: []Annotation{
				next(available(), mut(), error()),
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

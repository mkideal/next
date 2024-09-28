//go:generate go run github.com/gopherd/tools/cmd/docgen@v0.0.8 -ext .mdx -I ./ -o ../../website/docs/api/preview -level 0 -M "---" -M "pagination_prev: null" -M "pagination_next: null" -M "---"
package grammar

// @api(Grammar) represents the custom grammar for the next files.
//
// The grammar is used to define a subset of the next files. It can limit the features of the next code according
// by your requirements. The grammar is a JSON object that contains rules.
type Grammar struct {
	// @api(Grammar.Package) represents the [Package](#Grammar/Package) grammar rules for the package declaration.
	//
	// Example:
	//
	//	```json
	//	{
	//	  "Package": {
	//	  }
	//	}
	//	```
	//
	// See [Package](#Grammar/Package) for more details.
	Package Package

	Import struct {
		Off         bool // import declaration is off
		Annotations []Annotation
	}
	Const struct {
		Off         bool     // const declaration is off
		Types       []string // type names: bool, int, float, string
		Annotations []Annotation
	}
	Enum struct {
		Off         bool     // enum declaration is off
		Types       []string // type names: int, float, string
		Annotations []Annotation
		Member      struct {
			Annotations   []Annotation
			ValueRequired bool // value is required
			OffValueExpr  bool // value expression is off
			OffIota       bool // iota is off
			Validators    []Validator
		}
	}
	Struct struct {
		Off         bool // struct declaration is off
		Annotations []Annotation
		Field       struct {
			Annotations []Annotation
			Validators  []Validator
		}
	}
	Interface struct {
		Off         bool // interface declaration is off
		Annotations []Annotation
		Method      struct {
			Annotations []Annotation
			Validators  []Validator
			Param       struct {
				Annotations []Annotation
				Validators  []Validator
			}
		}
	}
}

// @api(Grammar/Validator) represents the validator for the grammar rules.
type Validator struct {
	// @api(Grammar/Validator.Name) represents the validator name.
	Name string

	// @api(Grammar/Validator.Expression) represents the validator expression.
	// The expression is a template string that can access the data by the `.` operator. The expression
	// must return a boolean value.
	//
	// The data is the current context object. For example, **package** object for the package validator.
	Expression string

	// @api(Grammar/Validator.Message) represents the error message when the validator is failed.
	Message string
}

// @api(Grammar/Annotation) represents the annotation grammar rules.
// Only the built-in annotations are supported if no annotations are defined in the grammar.
//
// :::warning
//
// You **CAN NOT** define the built-in annotations in the grammar.
//
// :::
type Annotation struct {
	// @api(Grammar/Annotation.Name) represents the annotation name.
	Name string

	// @api(Grammar/Annotation.Description) represents the annotation description.
	Description string

	// @api(Grammar/Annotation.Required) represents the annotation is required or not.
	Required bool

	// @api(Grammar/Annotation.Validators) represents the [Validator](#Grammar/Validator) for the annotation.
	Validators []Validator

	// @api(Grammar/Annotation.Parameters) represents the annotation parameters.
	Parameters []AnnotationParameter
}

// @api(Grammar/Annotation/Parameter) represents the annotation parameter grammar rules.
// If no parameters are defined, the annotation does not have any parameters.
type AnnotationParameter struct {
	// @api(Grammar/Annotation/Parameter.Name) represents the parameter name.
	Name string

	// @api(Grammar/Annotation/Parameter.Type) represents the parameter type.
	// The type is a string that can be one of the following types:
	//
	// - **bool**: boolean type, the value can be `true` or `false`
	// - **int**: integer type, the value can be a positive or negative integer, for example, `123`
	// - **float**: float type, the value can be a positive or negative float, for example, `1.23`
	// - **string**: string type, the value can be a string, for example, `"hello"`
	// - **type**: any type name, for example, `int`, `float`, `string`, etc. Custom type names are supported.
	Type string

	// @api(Grammar/Annotation/Parameter.Required) represents the parameter is required or not.
	Required bool

	// @api(Grammar/Annotation/Parameter.Validators) represents the [Validator](#Grammar/Validator) for the annotation parameter.
	Validators []Validator
}

// @api(Grammar/Package) represents the grammar rules for the package declaration.
type Package struct {
	// @api(Grammar/Package.Annotations) represents the [AnnotationGrammar](#Grammar/Annotation) rules for the package declaration.
	// It extends the built-in annotations.
	Annotations []Annotation

	// @api(Grammar/Package.Validators) represents the [Validator](#Grammar/Validator) for the package declaration.
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
	Validators []Validator
}

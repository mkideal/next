package ast

import (
	"github.com/gopherd/next/token"
)

// ----------------------------------------------------------------------------
// Objects

// An Object describes a named language entity such as a package,
// constant, type, variable, function (incl. methods), or label.
//
// The Data fields contains object-specific data:
//
//	Kind    Data type         Data value
//	Pkg     *Scope            package scope
//	Con     int               iota for the respective declaration
//
// Deprecated: The relationship between Idents and Objects cannot be
// correctly computed without type information. For example, the
// expression T{K: 0} may denote a struct, map, slice, or array
// literal, depending on the type of T. If T is a struct, then K
// refers to a field of T, whereas for the other types it refers to a
// value in the environment.
//
// New programs should set the [parser.SkipObjectResolution] parser
// flag to disable syntactic object resolution (which also saves CPU
// and memory), and instead use the type checker [go/types] if object
// resolution is desired. See the Defs, Uses, and Implicits fields of
// the [types.Info] struct for details.
type Object struct {
	Kind ObjKind
	Name string // declared name
	Decl any    // corresponding Field, XxxSpec, FuncDecl, LabeledStmt, AssignStmt, Scope; or nil
	Data any    // object-specific data; or nil
	Type any    // placeholder for type information; may be nil
}

// NewObj creates a new object of a given kind and name.
func NewObj(kind ObjKind, name string) *Object {
	return &Object{Kind: kind, Name: name}
}

// Pos computes the source position of the declaration of an object name.
// The result may be an invalid position if it cannot be computed
// (obj.Decl may be nil or not correct).
func (obj *Object) Pos() token.Pos {
	name := obj.Name
	switch d := obj.Decl.(type) {
	case *Field:
		if d.Name.Name == name {
			return d.Name.Pos()
		}
	case *ImportSpec:
		return d.Path.Pos()
	case *ValueSpec:
		if d.Name.Name == name {
			return d.Name.Pos()
		}
	case *TypeSpec:
		if d.Name.Name == name {
			return d.Name.Pos()
		}
	}
	return token.NoPos
}

// ObjKind describes what an [Object] represents.
type ObjKind int

// The list of possible [Object] kinds.
const (
	Bad ObjKind = iota // for error handling
	Pkg                // package
	Con                // constant
	Var                // variable
	Typ                // type
	Fun                // function or method
)

var objKindStrings = [...]string{
	Bad: "bad",
	Pkg: "package",
	Con: "const",
	Typ: "type",
	Fun: "func",
}

func (kind ObjKind) String() string { return objKindStrings[kind] }

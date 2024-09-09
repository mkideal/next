package ast

import (
	"fmt"
)

// Visitor represents an interface for traversing an AST.
// The Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// walkList applies the visitor v to each node in the list.
func walkList[N Node](v Visitor, list []N) {
	for _, node := range list {
		Walk(v, node)
	}
}

// walkGenDecl applies the visitor v to each part of a generic declaration.
func walkGenDecl[T Node](v Visitor, decl *GenDecl[T]) {
	if decl.Doc != nil {
		Walk(v, decl.Doc)
	}
	if decl.Annotations != nil {
		Walk(v, decl.Annotations)
	}
	Walk(v, decl.Name)
	Walk(v, decl.Spec)
	if decl.Comment != nil {
		Walk(v, decl.Comment)
	}
}

// Walk traverses an AST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	// Walk children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Comments and fields
	case *Comment:
		// nothing to do

	case *CommentGroup:
		walkList(v, n.List)

	case *AnnotationParam:
		Walk(v, n.Name)
		if n.Value != nil {
			Walk(v, n.Value)
		}

	case *Annotation:
		Walk(v, n.Name)
		walkList(v, n.Params)

	case *AnnotationGroup:
		walkList(v, n.List)

	case *StructField:
		if n.Doc != nil {
			Walk(v, n.Doc)
		}
		if n.Annotations != nil {
			Walk(v, n.Annotations)
		}
		Walk(v, n.Type)
		Walk(v, n.Name)
		if n.Comment != nil {
			Walk(v, n.Comment)
		}

	case *FieldList:
		walkList(v, n.List)

	case *MemberList:
		walkList(v, n.List)

	case *MethodList:
		walkList(v, n.List)

	case *Method:
		if n.Doc != nil {
			Walk(v, n.Doc)
		}
		if n.Annotations != nil {
			Walk(v, n.Annotations)
		}
		Walk(v, n.Name)
		Walk(v, n.Params)
		if n.Return != nil {
			Walk(v, n.Return)
		}
		if n.Comment != nil {
			Walk(v, n.Comment)
		}

	case *MethodParam:
		if n.Doc != nil {
			Walk(v, n.Doc)
		}
		if n.Annotations != nil {
			Walk(v, n.Annotations)
		}
		if n.Name != nil {
			Walk(v, n.Name)
		}
		Walk(v, n.Type)
		if n.Comment != nil {
			Walk(v, n.Comment)
		}

	case *MethodParamList:
		walkList(v, n.List)

	// Expressions
	case *BadExpr, *Ident, *BasicLit:
		// nothing to do

	case *ParenExpr:
		Walk(v, n.X)

	case *SelectorExpr:
		Walk(v, n.X)
		Walk(v, n.Sel)

	case *CallExpr:
		Walk(v, n.Fun)
		walkList(v, n.Args)

	case *UnaryExpr:
		Walk(v, n.X)

	case *BinaryExpr:
		Walk(v, n.X)
		Walk(v, n.Y)

	// Types
	case *ArrayType:
		Walk(v, n.T)
		Walk(v, n.N)

	case *VectorType:
		Walk(v, n.T)

	case *MapType:
		Walk(v, n.K)
		Walk(v, n.V)

	case *EnumType:
		Walk(v, n.Members)

	case *StructType:
		Walk(v, n.Fields)

	case *InterfaceType:
		Walk(v, n.Methods)

	// Declarations
	case *ImportDecl:
		if n.Doc != nil {
			Walk(v, n.Doc)
		}
		Walk(v, n.Path)
		if n.Comment != nil {
			Walk(v, n.Comment)
		}

	case *ConstDecl:
		walkGenDecl(v, n)

	case *EnumDecl:
		walkGenDecl(v, n)

	case *StructDecl:
		walkGenDecl(v, n)

	case *InterfaceDecl:
		walkGenDecl(v, n)

	case *EnumMember:
		if n.Doc != nil {
			Walk(v, n.Doc)
		}
		if n.Annotations != nil {
			Walk(v, n.Annotations)
		}
		Walk(v, n.Name)
		if n.Value != nil {
			Walk(v, n.Value)
		}
		if n.Comment != nil {
			Walk(v, n.Comment)
		}

	// Statements
	case *ExprStmt:
		Walk(v, n.X)

	// Files and packages
	case *File:
		if n.Doc != nil {
			Walk(v, n.Doc)
		}
		if n.Annotations != nil {
			Walk(v, n.Annotations)
		}
		Walk(v, n.Name)
		walkList(v, n.Decls)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}

	v.Visit(nil)
}

// inspector is a function type that determines whether to continue inspecting an AST.
type inspector func(Node) bool

// Visit implements the Visitor interface for the inspector type.
func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}

// Inspect traverses an AST in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Inspect invokes f
// recursively for each of the non-nil children of node, followed by a
// call of f(nil).
func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}

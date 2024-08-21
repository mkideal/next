package ast

import (
	"fmt"
	"iter"
)

// A Visitor's Visit method is invoked for each node encountered by [Walk].
// If the result visitor w is not nil, [Walk] visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node Node) (w Visitor)
}

func walkList[N Node](v Visitor, list []N) {
	for _, node := range list {
		Walk(v, node)
	}
}

// TODO(gri): Investigate if providing a closure to Walk leads to
// simpler use (and may help eliminate Inspect in turn).

// Walk traverses an AST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Comments and fields
	case *Comment:
		// nothing to do

	case *CommentGroup:
		walkList(v, n.List)

	case *AnnotationParam:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.Value != nil {
			Walk(v, n.Value)
		}

	case *Annotation:
		Walk(v, n.Name)
		walkList(v, n.Params)

	case *AnnotationGroup:
		walkList(v, n.List)

	case *Field:
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

	case *StructType:
		Walk(v, n.Fields)

	case *EnumType:
		walkList(v, n.Values)

	// Declarations
	case *ImportSpec:
		if n.Doc != nil {
			Walk(v, n.Doc)
		}
		if n.Annotations != nil {
			Walk(v, n.Annotations)
		}
		Walk(v, n.Path)
		if n.Comment != nil {
			Walk(v, n.Comment)
		}

	case *ValueSpec:
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

	case *TypeSpec:
		if n.Doc != nil {
			Walk(v, n.Doc)
		}
		if n.Annotations != nil {
			Walk(v, n.Annotations)
		}
		Walk(v, n.Name)
		Walk(v, n.Type)
		if n.Comment != nil {
			Walk(v, n.Comment)
		}

	case *BadDecl:
		// nothing to do

	case *GenDecl:
		if n.Doc != nil {
			Walk(v, n.Doc)
		}
		if n.Annotations != nil {
			Walk(v, n.Annotations)
		}
		walkList(v, n.Specs)

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

	case *Package:
		for _, f := range n.Files {
			Walk(v, f)
		}

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}

	v.Visit(nil)
}

type inspector func(Node) bool

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

// Preorder returns an iterator over all the nodes of the syntax tree
// beneath (and including) the specified root, in depth-first
// preorder.
//
// For greater control over the traversal of each subtree, use [Inspect].
func Preorder(root Node) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		ok := true
		Inspect(root, func(n Node) bool {
			if n != nil {
				// yield must not be called once ok is false.
				ok = ok && yield(n)
			}
			return ok
		})
	}
}

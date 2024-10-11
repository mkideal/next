package compile

import (
	"github.com/mkideal/next/src/ast"
	"github.com/mkideal/next/src/constant"
	"github.com/mkideal/next/src/token"
)

// Stmt represents a Next statement.
type Stmt interface {
	stmtNode()
	resolve(*Compiler, *File)
}

func newStmt(c *Compiler, file *File, src ast.Stmt) Stmt {
	switch s := src.(type) {
	case *ast.ExprStmt:
		x := ast.Unparen(s.X)
		if call, ok := x.(*ast.CallExpr); ok {
			return newCallStmt(c, file, call)
		}
	}
	c.addErrorf(src.Pos(), "unsupported statement: %T", src)
	return nil
}

type CallStmt struct {
	pos      token.Pos
	CallExpr *ast.CallExpr
}

func newCallStmt(c *Compiler, file *File, call *ast.CallExpr) *CallStmt {
	s := &CallStmt{pos: call.Pos(), CallExpr: call}
	file.addObject(c, call, s)
	return s
}

func (s *CallStmt) resolve(c *Compiler, file *File) {
	fun := ast.Unparen(s.CallExpr.Fun)
	ident, ok := fun.(*ast.Ident)
	if !ok {
		c.addErrorf(fun.Pos(), "unexpected function %T", fun)
		return
	}
	args := make([]constant.Value, len(s.CallExpr.Args))
	for i, arg := range s.CallExpr.Args {
		args[i] = c.resolveValue(file, arg, nil)
	}
	_, err := c.call(s.pos, s.CallExpr.Fun, args...)
	if err != nil {
		c.addErrorf(s.CallExpr.Pos(), "call %s: %s", ident.Name, err)
	}
}

func (s *CallStmt) stmtNode() {}

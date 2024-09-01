package types

import (
	"github.com/next/next/src/ast"
	"github.com/next/next/src/constant"
	"github.com/next/next/src/token"
)

// Stmt represents a Next statement.
type Stmt interface {
	stmtNode()
	resolve(ctx *Context, file *File)
}

func newStmt(ctx *Context, file *File, src ast.Stmt) Stmt {
	switch s := src.(type) {
	case *ast.ExprStmt:
		x := ast.Unparen(s.X)
		if call, ok := x.(*ast.CallExpr); ok {
			return newCallStmt(ctx, file, call)
		}
	}
	ctx.addErrorf(src.Pos(), "unsupported statement: %T", src)
	return nil
}

type CallStmt struct {
	pos      token.Pos
	CallExpr *ast.CallExpr
}

func newCallStmt(_ *Context, _ *File, call *ast.CallExpr) *CallStmt {
	return &CallStmt{pos: call.Pos(), CallExpr: call}
}

func (s *CallStmt) resolve(ctx *Context, file *File) {
	fun := ast.Unparen(s.CallExpr.Fun)
	ident, ok := fun.(*ast.Ident)
	if !ok {
		ctx.addErrorf(fun.Pos(), "unexpected function %T", fun)
		return
	}
	args := make([]constant.Value, len(s.CallExpr.Args))
	for i, arg := range s.CallExpr.Args {
		args[i] = ctx.resolveValue(file, arg, nil)
	}
	_, err := ctx.call(s.pos, ident.Name, args...)
	if err != nil {
		ctx.addErrorf(s.CallExpr.Pos(), "call %s: %s", ident.Name, err)
	}
}

func (s *CallStmt) stmtNode() {}
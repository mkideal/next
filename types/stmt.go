package types

import (
	"github.com/gopherd/next/ast"
	"github.com/gopherd/next/constant"
	"github.com/gopherd/next/token"
)

type Stmt interface {
	Node
	stmtNode()
	resolve(ctx *Context, file *File)
}

func newStmt(ctx *Context, src ast.Stmt) Stmt {
	switch s := src.(type) {
	case *ast.ExprStmt:
		x := ast.Unparen(s.X)
		if call, ok := x.(*ast.CallExpr); ok {
			return newCallStmt(ctx, call)
		}
	}
	ctx.errorf(src.Pos(), "unsupported statement: %T", src)
	return nil
}

type CallStmt struct {
	pos      token.Pos
	CallExpr *ast.CallExpr
}

func newCallStmt(ctx *Context, call *ast.CallExpr) *CallStmt {
	return &CallStmt{pos: call.Pos(), CallExpr: call}
}

func (s *CallStmt) resolve(ctx *Context, file *File) {
	fun := ast.Unparen(s.CallExpr.Fun)
	ident, ok := fun.(*ast.Ident)
	if !ok {
		ctx.errorf(fun.Pos(), "unexpected function %T", fun)
		return
	}
	args := make([]constant.Value, len(s.CallExpr.Args))
	for i, arg := range s.CallExpr.Args {
		args[i] = ctx.resolveValue(file, arg, nil)
	}
	_, err := ctx.call(s.pos, ident.Name, args...)
	if err != nil {
		ctx.errorf(s.CallExpr.Pos(), "call %s: %s", ident.Name, err)
	}
}

func (s *CallStmt) Pos() token.Pos { return s.pos }
func (s *CallStmt) stmtNode()      {}

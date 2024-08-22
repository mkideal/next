package constant

import (
	"fmt"
	"io"

	"github.com/gopherd/next/token"
)

type FuncContext interface {
	Output() io.Writer
	Debug() bool
	Position() token.Position
}

type Func func(ctx FuncContext, args []Value) Value

var funcs = map[string]Func{
	"len":       _len,
	"min":       _min,
	"max":       _max,
	"abs":       _abs,
	"int":       _int,
	"float":     _float,
	"bool":      _bool,
	"sprint":    _sprint,
	"sprintf":   _sprintf,
	"sprintln":  _sprintln,
	"print":     _print,
	"printf":    _printf,
	"println":   _println,
	"error":     _error,
	"assert":    _assert,
	"assert_eq": _assert_eq,
	//"assert_ne": _assert_ne,
	//"assert_lt": _assert_lt,
	//"assert_le": _assert_le,
	//"assert_gt": _assert_gt,
	//"assert_ge": _assert_ge,
}

func _len(ctx FuncContext, args []Value) Value {
	if len(args) != 1 {
		return unknownVal{}
	}
	if v := args[0]; v.Kind() == String {
		return MakeUint64(uint64(len(StringVal(v))))
	}
	return unknownVal{}
}

func _min(ctx FuncContext, args []Value) Value {
	if len(args) < 1 {
		panic("min: missing arguments")
	}
	min := args[0]
	if min.Kind() == Unknown {
		return min
	}
	for _, v := range args[1:] {
		if v.Kind() == Unknown {
			return v
		}
		if Compare(v, token.LSS, min) {
			min = v
		}
	}
	return min
}

func _max(ctx FuncContext, args []Value) Value {
	if len(args) < 1 {
		panic("max: missing arguments")
	}
	max := args[0]
	if max.Kind() == Unknown {
		return max
	}
	for _, v := range args[1:] {
		if v.Kind() == Unknown {
			return v
		}
		if Compare(v, token.GTR, max) {
			max = v
		}
	}
	return max
}

func _abs(ctx FuncContext, args []Value) Value {
	if len(args) != 1 {
		panic("abs: exactly one argument is required")
	}
	v := args[0]
	if v.Kind() == Unknown {
		return v
	}
	switch v.Kind() {
	case Int:
		if i, ok := Int64Val(v); ok {
			if i < 0 {
				return MakeInt64(-i)
			}
			return v
		} else if _, ok := Uint64Val(v); ok {
			return v
		}
	case Float:
		if f, ok := Float64Val(v); ok {
			if f < 0 {
				return MakeFloat64(-f)
			}
			return v
		}
	}
	return MakeUnknown()
}

func _int(ctx FuncContext, args []Value) Value {
	if len(args) != 1 {
		panic("int: exactly one argument is required")
	}
	v := args[0]
	if v.Kind() == Unknown {
		return v
	}
	switch v.Kind() {
	case Bool:
		if BoolVal(v) {
			return MakeInt64(1)
		}
		return MakeInt64(0)
	case Int:
		return v
	case Float:
		f, _ := Float64Val(v)
		return MakeInt64(int64(f))
	}
	return MakeUnknown()
}

func _float(ctx FuncContext, args []Value) Value {
	if len(args) != 1 {
		panic("int: exactly one argument is required")
	}
	v := args[0]
	if v.Kind() == Unknown {
		return MakeUnknown()
	}
	switch v.Kind() {
	case Bool:
		if BoolVal(v) {
			return MakeFloat64(1)
		}
		return MakeFloat64(0)
	case Int:
		if i, ok := Int64Val(v); ok {
			return MakeFloat64(float64(i))
		} else if i, ok := Uint64Val(v); ok {
			return MakeFloat64(float64(i))
		}
	case Float:
		return v
	}
	return MakeUnknown()
}

func _bool(ctx FuncContext, args []Value) Value {
	if len(args) != 1 {
		panic("int: exactly one argument is required")
	}
	v := args[0]
	if v.Kind() == Unknown {
		return MakeUnknown()
	}
	switch v.Kind() {
	case Bool:
		return v
	case Int:
		if i, ok := Int64Val(v); ok {
			return MakeBool(i != 0)
		} else if i, ok := Uint64Val(v); ok {
			return MakeBool(i != 0)
		}
	case Float:
		f, _ := Float64Val(v)
		return MakeBool(f != 0)
	case String:
		return MakeBool(len(StringVal(v)) > 0)
	}
	return MakeUnknown()
}

func toArgs(args []Value) []any {
	if len(args) == 0 {
		return nil
	}
	argv := make([]any, len(args))
	for i, v := range args {
		switch v.Kind() {
		case Bool:
			argv[i] = BoolVal(v)
		case Int:
			argv[i], _ = Int64Val(v)
		case Float:
			argv[i], _ = Float64Val(v)
		case String:
			argv[i] = StringVal(v)
		default:
			argv[i] = nil
		}
	}
	return argv
}

func _sprint(ctx FuncContext, args []Value) Value {
	return MakeString(fmt.Sprint(toArgs(args)...))
}

func _sprintf(ctx FuncContext, args []Value) Value {
	if len(args) == 0 {
		panic("sprintf: missing format string")
	}
	format := args[0]
	if format.Kind() != String {
		return MakeUnknown()
	}
	return MakeString(fmt.Sprintf(StringVal(format), toArgs(args[1:])...))
}

func _sprintln(ctx FuncContext, args []Value) Value {
	return MakeString(fmt.Sprintln(toArgs(args)...))
}

func _print(ctx FuncContext, args []Value) Value {
	if !ctx.Debug() {
		return MakeUnknown()
	}
	fmt.Fprint(ctx.Output(), toArgs(args)...)
	return MakeUnknown()
}

func _printf(ctx FuncContext, args []Value) Value {
	if !ctx.Debug() {
		return MakeUnknown()
	}
	if len(args) == 0 {
		panic("printf: missing format string")
	}
	format := args[0]
	if format.Kind() != String {
		fmt.Fprintln(ctx.Output(), "printf: format string is not a string")
		return MakeUnknown()
	}
	fmt.Fprintf(ctx.Output(), StringVal(format), toArgs(args[1:])...)
	return MakeUnknown()
}

func _println(ctx FuncContext, args []Value) Value {
	if !ctx.Debug() {
		return MakeUnknown()
	}
	fmt.Fprintln(ctx.Output(), toArgs(args)...)
	return MakeUnknown()
}

func _error(ctx FuncContext, args []Value) Value {
	if len(args) == 0 {
		panic("error: missing error message")
	}
	fmt.Fprintln(ctx.Output(), toArgs(args)...)
	return MakeUnknown()
}

func printAssert(ctx FuncContext, args []any) {
	if len(args) == 0 {
		fmt.Fprintf(ctx.Output(), "%s: assertion failed\n", ctx.Position())
		return
	}
	fmt.Fprintf(ctx.Output(), "%s: assertion failed: %s\n", ctx.Position(), fmt.Sprint(args...))
}

func _assert(ctx FuncContext, args []Value) Value {
	if len(args) == 0 {
		panic("assert: missing condition")
	}
	if !BoolVal(args[0]) {
		printAssert(ctx, toArgs(args[1:]))
	}
	return MakeUnknown()
}

func _assert_eq(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_eq: at least two arguments are required")
	}
	if !Compare(args[0], token.EQL, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v, but got %v.", args[1], args[0])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

func _assert_ne(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_ne: at least two arguments are required")
	}
	if Compare(args[0], token.EQL, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected not %v, but got %v.", args[1], args[0])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

func _assert_lt(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_lt: at least two arguments are required")
	}
	if !Compare(args[0], token.LSS, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v < %v, but not", args[0], args[1])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

func _assert_le(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_le: at least two arguments are required")
	}
	if !Compare(args[0], token.LEQ, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v <= %v, but not", args[0], args[1])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

func _assert_gt(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_gt: at least two arguments are required")
	}
	if !Compare(args[0], token.GTR, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v > %v, but not", args[0], args[1])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

func _assert_ge(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_ge: at least two arguments are required")
	}
	if !Compare(args[0], token.GEQ, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v >= %v, but not", args[0], args[1])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

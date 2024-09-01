// Package constant provides functionality for working with constant values and functions.
// It includes a set of predefined functions that can be called dynamically,
// as well as utility functions for type conversion and assertion.
package constant

import (
	"fmt"

	"github.com/next/next/src/token"
)

// FuncContext represents the context in which a function is executed.
// It provides methods for debugging, printing, and error reporting.
type FuncContext interface {
	IsDebugEnabled() bool
	Print(args ...any)
	Error(args ...any)
}

// Func represents a function that can be called with a FuncContext and arguments.
type Func func(ctx FuncContext, args []Value) Value

// Call invokes the function named by fun with the provided arguments.
// It returns the result of the function call and any error that occurred.
func Call(ctx FuncContext, fun string, args []Value) (v Value, err error) {
	f, ok := funcs[fun]
	if !ok {
		return unknownVal{}, fmt.Errorf("unknown function %q", fun)
	}
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case error:
				err = e
			case string:
				err = fmt.Errorf("failed to call %q: %s", fun, e)
			default:
				panic(r)
			}
		}
	}()
	return f(ctx, args), nil
}

// funcs is a map of predefined functions that can be called using the Call function.
var funcs = map[string]Func{
	"len":       lenFunc,
	"min":       minFunc,
	"max":       maxFunc,
	"abs":       absFunc,
	"int":       intFunc,
	"float":     floatFunc,
	"bool":      boolFunc,
	"sprint":    sprintFunc,
	"sprintf":   sprintfFunc,
	"sprintln":  sprintlnFunc,
	"print":     printFunc,
	"printf":    printfFunc,
	"error":     errorFunc,
	"assert":    assertFunc,
	"assert_eq": assertEqFunc,
	"assert_ne": assertNeFunc,
	"assert_lt": assertLtFunc,
	"assert_le": assertLeFunc,
	"assert_gt": assertGtFunc,
	"assert_ge": assertGeFunc,
}

// lenFunc returns the length of a string.
func lenFunc(ctx FuncContext, args []Value) Value {
	if len(args) != 1 {
		return unknownVal{}
	}
	if v := args[0]; v.Kind() == String {
		return MakeUint64(uint64(len(StringVal(v))))
	}
	return unknownVal{}
}

// minFunc returns the minimum value from the provided arguments.
func minFunc(ctx FuncContext, args []Value) Value {
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

// maxFunc returns the maximum value from the provided arguments.
func maxFunc(ctx FuncContext, args []Value) Value {
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

// absFunc returns the absolute value of the provided argument.
func absFunc(ctx FuncContext, args []Value) Value {
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

// intFunc converts the provided argument to an integer value.
func intFunc(ctx FuncContext, args []Value) Value {
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

// floatFunc converts the provided argument to a float value.
func floatFunc(ctx FuncContext, args []Value) Value {
	if len(args) != 1 {
		panic("float: exactly one argument is required")
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

// boolFunc converts the provided argument to a boolean value.
func boolFunc(ctx FuncContext, args []Value) Value {
	if len(args) != 1 {
		panic("bool: exactly one argument is required")
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

// toArgs converts a slice of Values to a slice of interface{}.
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

// sprintFunc returns a string representation of the provided arguments.
func sprintFunc(ctx FuncContext, args []Value) Value {
	return MakeString(fmt.Sprint(toArgs(args)...))
}

// sprintfFunc returns a formatted string using the provided format and arguments.
func sprintfFunc(ctx FuncContext, args []Value) Value {
	if len(args) == 0 {
		panic("sprintf: missing format string")
	}
	format := args[0]
	if format.Kind() != String {
		return MakeUnknown()
	}
	return MakeString(fmt.Sprintf(StringVal(format), toArgs(args[1:])...))
}

// sprintlnFunc returns a string representation of the provided arguments, followed by a newline.
func sprintlnFunc(ctx FuncContext, args []Value) Value {
	return MakeString(fmt.Sprintln(toArgs(args)...))
}

// printFunc prints the provided arguments if debugging is enabled.
func printFunc(ctx FuncContext, args []Value) Value {
	if !ctx.IsDebugEnabled() {
		return MakeUnknown()
	}
	ctx.Print(fmt.Sprint(toArgs(args)...))
	return MakeUnknown()
}

// printfFunc prints a formatted string using the provided format and arguments if debugging is enabled.
func printfFunc(ctx FuncContext, args []Value) Value {
	if !ctx.IsDebugEnabled() {
		return MakeUnknown()
	}
	if len(args) == 0 {
		panic("printf: missing format string")
	}
	format := args[0]
	if format.Kind() != String {
		panic("printf: format string is not a string")
	}
	ctx.Print(fmt.Sprintf(StringVal(format), toArgs(args[1:])...))
	return MakeUnknown()
}

// errorFunc reports an error with the provided message.
func errorFunc(ctx FuncContext, args []Value) Value {
	if len(args) == 0 {
		panic("error: missing error message")
	}
	ctx.Error(fmt.Sprint(toArgs(args)...))
	return MakeUnknown()
}

// printAssert prints an assertion failure message.
func printAssert(ctx FuncContext, args []any) {
	if len(args) == 0 {
		ctx.Error("assertion failed")
		return
	}
	ctx.Error("assertion failed: " + fmt.Sprint(args...))
}

// assertFunc checks if the provided condition is true and reports an error if it's not.
func assertFunc(ctx FuncContext, args []Value) Value {
	if len(args) == 0 {
		panic("assert: missing condition")
	}
	if !BoolVal(args[0]) {
		printAssert(ctx, toArgs(args[1:]))
	}
	return MakeUnknown()
}

// assertEqFunc checks if two values are equal and reports an error if they're not.
func assertEqFunc(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_eq: at least two arguments are required")
	}
	if !Compare(args[0], token.EQL, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v, but got %v", args[1], args[0])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

// assertNeFunc checks if two values are not equal and reports an error if they are.
func assertNeFunc(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_ne: at least two arguments are required")
	}
	if Compare(args[0], token.EQL, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected not %v, but got %v", args[1], args[0])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

// assertLtFunc checks if the first value is less than the second and reports an error if it's not.
func assertLtFunc(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_lt: at least two arguments are required")
	}
	if !Compare(args[0], token.LSS, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v < %v, but not", args[0], args[1])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

// assertLeFunc checks if the first value is less than or equal to the second and reports an error if it's not.
func assertLeFunc(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_le: at least two arguments are required")
	}
	if !Compare(args[0], token.LEQ, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v <= %v, but not", args[0], args[1])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

// assertGtFunc checks if the first value is greater than the second and reports an error if it's not.
func assertGtFunc(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_gt: at least two arguments are required")
	}
	if !Compare(args[0], token.GTR, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v > %v, but not", args[0], args[1])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

// assertGeFunc checks if the first value is greater than or equal to the second and reports an error if it's not.
func assertGeFunc(ctx FuncContext, args []Value) Value {
	if len(args) < 2 {
		panic("assert_ge: at least two arguments are required")
	}
	if !Compare(args[0], token.GEQ, args[1]) {
		printAssert(ctx, append([]any{fmt.Sprintf("expected %v >= %v, but not", args[0], args[1])}, toArgs(args[2:])...))
	}
	return MakeUnknown()
}

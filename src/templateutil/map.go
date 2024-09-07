package templateutil

import (
	"fmt"
	"reflect"
)

// Func is a function that converts a reflect.Value to another reflect.Value.
type Func func(reflect.Value) (reflect.Value, error)

// Func2 is a function that converts a T and a reflect.Value to another reflect.Value.
type Func2[T any] func(T, reflect.Value) (reflect.Value, error)

// Func3 is a function that converts a T1, T2 and a reflect.Value to another reflect.Value.
type Func3[T1, T2 any] func(T1, T2, reflect.Value) (reflect.Value, error)

// Func4 is a function that converts a T1, T2, T3 and a reflect.Value to another reflect.Value.
type Func4[T1, T2, T3 any] func(T1, T2, T3, reflect.Value) (reflect.Value, error)

// FuncChain is a function chain that can be used to convert values.
//
// The function must accept 0 or 1 arguments.
//
// If the function accepts 0 arguments, it returns itself (the function).
// If the function accepts 1 argument and the argument is a FuncChain, it returns a new FuncChain
// that chains the two functions.
// Otherwise, it returns the result of calling the function with the argument.
type FuncChain func(...reflect.Value) (reflect.Value, error)

// Chain returns a FuncChain that chains the given function.
func Chain(f Func) FuncChain {
	var self FuncChain
	self = FuncChain(func(values ...reflect.Value) (reflect.Value, error) {
		return call(self, f, values...)
	})
	return self
}

// Chain2 returns a FuncChain that chains the given function.
func Chain2[T any](f Func2[T]) func(T, ...reflect.Value) (reflect.Value, error) {
	return func(x T, argument ...reflect.Value) (reflect.Value, error) {
		var self FuncChain
		self = FuncChain(func(arg ...reflect.Value) (reflect.Value, error) {
			return call(self, func(y reflect.Value) (reflect.Value, error) {
				return f(x, y)
			}, arg...)
		})
		return self(argument...)
	}
}

// Chain3 returns a FuncChain that chains the given function.
func Chain3[T1, T2 any](f Func3[T1, T2]) func(T1, T2, ...reflect.Value) (reflect.Value, error) {
	return func(x T1, y T2, argument ...reflect.Value) (reflect.Value, error) {
		var self FuncChain
		self = FuncChain(func(arg ...reflect.Value) (reflect.Value, error) {
			return call(self, func(z reflect.Value) (reflect.Value, error) {
				return f(x, y, z)
			}, arg...)
		})
		return self(argument...)
	}
}

// Chain4 returns a FuncChain that chains the given function.
func Chain4[T1, T2, T3 any](f Func4[T1, T2, T3]) func(T1, T2, T3, ...reflect.Value) (reflect.Value, error) {
	return func(x T1, y T2, z T3, argument ...reflect.Value) (reflect.Value, error) {
		var self FuncChain
		self = FuncChain(func(arg ...reflect.Value) (reflect.Value, error) {
			return call(self, func(w reflect.Value) (reflect.Value, error) {
				return f(x, y, z, w)
			}, arg...)
		})
		return self(argument...)
	}
}

// call calls the given function with the given arguments in the function chain.
// If no arguments are given, it returns the function chain itself.
// If one argument is given and it is a FuncChain, it returns a new FuncChain that
// chains the two functions.
// Otherwise, it returns the result of calling the function with the argument.
func call(fc FuncChain, f Func, argument ...reflect.Value) (reflect.Value, error) {
	if len(argument) == 0 {
		// Return self if no arguments are given.
		return reflect.ValueOf(fc), nil
	}
	if len(argument) > 1 {
		return reflect.Value{}, fmt.Errorf("expected 0 or 1 arguments, got %d", len(argument))
	}
	v := argument[0]
	if v.Kind() != reflect.Func {
		return f(v)
	}
	if !v.CanInterface() {
		return reflect.Value{}, fmt.Errorf("function is not exported")
	}
	c, ok := v.Interface().(FuncChain)
	if !ok {
		return reflect.Value{}, fmt.Errorf("expected function chain, got %s", v.Type())
	}
	return reflect.ValueOf(c.then(f)), nil
}

// then returns a new function chain that chains the given function with the current function chain.
func (f FuncChain) then(next Func) FuncChain {
	return Chain(func(v reflect.Value) (reflect.Value, error) {
		v, err := f(v)
		if err != nil {
			return reflect.Value{}, err
		}
		return next(v)
	})
}

// Map maps the given value to a slice of values using the given function chain.
func Map(f FuncChain, v Slice) (Slice, error) {
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return null, fmt.Errorf("map: expected slice or array, got %s", v.Type())
	}
	if v.Len() == 0 {
		return reflect.MakeSlice(reflect.SliceOf(v.Type().Elem()), 0, 0), nil
	}
	var result reflect.Value
	for i := 0; i < v.Len(); i++ {
		r, err := f(v.Index(i))
		if err != nil {
			return null, err
		}
		if i == 0 {
			result = reflect.MakeSlice(reflect.SliceOf(r.Type()), v.Len(), v.Len())
		}
		result.Index(i).Set(r)
	}
	return result, nil
}

// noError returns a function that calls the given function and returns the result and nil.
func noError[T, U any](f func(T) U) func(T) (U, error) {
	return func(s T) (U, error) {
		return f(s), nil
	}
}

// stringFunc converts a function that takes a string and returns a string to a funtion
// that takes a reflect.Value and returns a reflect.Value.
func stringFunc(name string, f func(string) (string, error)) Func {
	return func(v reflect.Value) (reflect.Value, error) {
		s, ok := asString(v)
		if !ok {
			return reflect.Value{}, fmt.Errorf("%s: expected string, got %s", name, v.Type())
		}
		r, err := f(s)
		return reflect.ValueOf(r), err
	}
}

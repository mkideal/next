package templateutil

import (
	"fmt"
	"reflect"
)

type ConverterFunc func(...reflect.Value) (reflect.Value, error)

func noError[T, U any](f func(T) U) func(T) (U, error) {
	return func(s T) (U, error) {
		return f(s), nil
	}
}

func noError2[T1, T2, U any](f func(T1, T2) U) func(T1, T2) (U, error) {
	return func(t1 T1, t2 T2) (U, error) {
		return f(t1, t2), nil
	}
}

func stringFunc(f func(string) (string, error)) func(reflect.Value) (reflect.Value, error) {
	return func(v reflect.Value) (reflect.Value, error) {
		s, ok := asString(v)
		if !ok {
			return reflect.Value{}, fmt.Errorf("expected string, got %s", v.Type())
		}
		r, err := f(s)
		return reflect.ValueOf(r), err
	}
}

func stringFunc2(f func(string, string) (string, error)) func(reflect.Value, reflect.Value) (reflect.Value, error) {
	return func(v1, v2 reflect.Value) (reflect.Value, error) {
		s1, ok := asString(v1)
		if !ok {
			return reflect.Value{}, fmt.Errorf("expected first argument to be string, got %s", v1.Type())
		}
		s2, ok := asString(v2)
		if !ok {
			return reflect.Value{}, fmt.Errorf("expected second argument to be string, got %s", v2.Type())
		}
		r, err := f(s1, s2)
		return reflect.ValueOf(r), err
	}
}

// conv returns a ConverterFunc that converts a value to another value using the given function.
func conv(f func(reflect.Value) (reflect.Value, error)) ConverterFunc {
	var self ConverterFunc
	self = ConverterFunc(func(values ...reflect.Value) (reflect.Value, error) {
		return doConvert(self, f, values...)
	})
	return self
}

func doConvert(self ConverterFunc, f func(reflect.Value) (reflect.Value, error), values ...reflect.Value) (reflect.Value, error) {
	if len(values) == 0 {
		// Return self if no arguments are given.
		return reflect.ValueOf(self), nil
	}
	if len(values) > 1 {
		return reflect.Value{}, fmt.Errorf("expected 0 or 1 arguments, got %d", len(values))
	}
	v := values[0]
	c, err := asConverterFunc(v)
	if err != nil {
		return reflect.Value{}, err
	}
	if c != nil {
		return reflect.ValueOf(c.thenAny(f)), nil
	}
	return f(v)
}

func conv2[T any](f func(T, reflect.Value) (reflect.Value, error)) func(T, ...reflect.Value) (reflect.Value, error) {
	return func(x T, values ...reflect.Value) (reflect.Value, error) {
		var self ConverterFunc
		self = ConverterFunc(func(values ...reflect.Value) (reflect.Value, error) {
			return doConvert(self, func(y reflect.Value) (reflect.Value, error) {
				return f(x, y)
			}, values...)
		})
		return self(values...)
	}
}

func conv3[T1, T2 any](f func(T1, T2, reflect.Value) (reflect.Value, error)) func(T1, T2, ...reflect.Value) (reflect.Value, error) {
	return func(x T1, y T2, values ...reflect.Value) (reflect.Value, error) {
		var self ConverterFunc
		self = ConverterFunc(func(values ...reflect.Value) (reflect.Value, error) {
			return doConvert(self, func(z reflect.Value) (reflect.Value, error) {
				return f(x, y, z)
			}, values...)
		})
		return self(values...)
	}
}

// convs returns a ConverterFunc that converts a string to another string using the given function.
func convs(f func(string) (string, error)) ConverterFunc {
	return conv(stringFunc(f))
}

func conv2s(f func(string, string) (string, error)) func(reflect.Value, ...reflect.Value) (reflect.Value, error) {
	return conv2(stringFunc2(f))
}

func asConverterFunc(v reflect.Value) (ConverterFunc, error) {
	if v.Kind() != reflect.Func {
		return nil, nil
	}
	c, ok := v.Interface().(ConverterFunc)
	if !ok {
		return nil, fmt.Errorf("expected ConverterFunc, got %s", v.Type())
	}
	return c, nil
}

func (f ConverterFunc) value() reflect.Value {
	return reflect.ValueOf(f)
}

func (f ConverterFunc) Convert(v reflect.Value) (reflect.Value, error) {
	v, err := f(v)
	if err != nil {
		return reflect.Value{}, err
	}
	return v, nil
}

func (f ConverterFunc) then(next func(string) (string, error)) ConverterFunc {
	return convs(func(s string) (string, error) {
		v, err := f.Convert(reflect.ValueOf(s))
		if err != nil {
			return "", err
		}
		s, ok := asString(v)
		if !ok {
			return "", fmt.Errorf("expected string, got %s", v.Type())
		}
		return next(s)
	})
}

func (f ConverterFunc) thenAny(next func(reflect.Value) (reflect.Value, error)) ConverterFunc {
	return conv(func(v reflect.Value) (reflect.Value, error) {
		v, err := f.Convert(v)
		if err != nil {
			return reflect.Value{}, err
		}
		return next(v)
	})
}

func mapFunc(c ConverterFunc, v reflect.Value) ([]string, error) {
	ss, err := toStrings(v)
	if err != nil {
		return nil, err
	}
	r := make([]string, 0, len(ss))
	if len(ss) == 0 {
		return r, nil
	}
	for _, s := range ss {
		v, err := c.Convert(reflect.ValueOf(s))
		if err != nil {
			return nil, err
		}
		if s, ok := asString(v); !ok {
			return nil, fmt.Errorf("expected string, got %s", v.Type())
		} else {
			r = append(r, s)
		}
	}
	return r, nil
}

func compositeConverters(converters ...ConverterFunc) (ConverterFunc, error) {
	c := converters[len(converters)-1]
	for i := len(converters) - 2; i >= 0; i-- {
		r, err := c(reflect.ValueOf(converters[i]))
		if err != nil {
			return nil, err
		}
		if r.Kind() != reflect.Func {
			return nil, fmt.Errorf("expected func, got %s", r.Type())
		}
		c = r.Interface().(ConverterFunc)
	}
	return c, nil
}

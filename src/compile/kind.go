package compile

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mkideal/next/src/grammar"
)

// @api(Object/Common/Type/Kinds) represents the type kind set.
type Kinds uint64

func (ks Kinds) contains(k Kind) bool {
	k = 1 << k
	return ks&Kinds(k) == Kinds(k)
}

// @api(Object/Common/Type/Kinds.Has) reports whether the type contains specific kind.
// The kind can be a `Kind` (or any integer) or a string representation of the [Kind](#Object/Common/Type/Kind).
// If the kind is invalid, it returns an error.
func (ks Kinds) Has(k any) (bool, error) {
	switch k := k.(type) {
	case Kind:
		return ks.contains(k), nil
	case string:
		var o Kind
		if err := o.Set(k); err != nil {
			return false, err
		}
		return ks.contains(o), nil
	default:
		switch reflect.TypeOf(k).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return ks.contains(Kind(reflect.ValueOf(k).Int())), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return ks.contains(Kind(reflect.ValueOf(k).Uint())), nil
		default:
			return false, fmt.Errorf("invalid type %T", k)
		}
	}
}

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=Kind -linecomment
type Kind int

func (k Kind) kinds() Kinds { return 1 << k }

func (k Kind) grammarType() string {
	if k.IsInteger() {
		return grammar.Int
	}
	if k.IsFloat() {
		return grammar.Float
	}
	if k.IsString() {
		return grammar.String
	}
	if k.IsBool() {
		return grammar.Bool
	}
	return ""
}

// @api(Object/Common/Type/Kind) represents the type kind.
// Currently, the following kinds are supported:
//
// - **bool**: true or false
// - **int**: integer
// - **int8**: 8-bit integer
// - **int16**: 16-bit integer
// - **int32**: 32-bit integer
// - **int64**: 64-bit integer
// - **float32**: 32-bit floating point
// - **float64**: 64-bit floating point
// - **byte**: byte
// - **bytes**: byte slice
// - **string**: string
// - **time**: time
// - **duration**: duration
// - **any**: any object
// - **map**: dictionary
// - **vector**: vector of elements
// - **array**: array of elements
// - **enum**: enumeration
// - **struct**: structure
// - **interface**: interface
const (
	KindInvalid   Kind = iota // invalid
	KindBool                  // bool
	KindInt                   // int
	KindInt8                  // int8
	KindInt16                 // int16
	KindInt32                 // int32
	KindInt64                 // int64
	KindFloat32               // float32
	KindFloat64               // float64
	KindByte                  // byte
	KindBytes                 // bytes
	KindString                // string
	KindTime                  // time
	KindDuration              // duration
	KindAny                   // any
	KindMap                   // map
	KindVector                // vector
	KindArray                 // array
	KindEnum                  // enum
	KindStruct                // struct
	KindInterface             // interface

	kindCount // -count-
)

// @api(Object/Common/Type/Kind.String) returns the string representation of the kind.

// @api(Object/Common/Type/Kind.Valid) reports whether the type is valid.
func (k Kind) Valid() bool {
	return k > KindInvalid && k < kindCount
}

func (k *Kind) Set(s string) error {
	switch s = strings.ToLower(s); s {
	case "bool":
		*k = KindBool
	case "int":
		*k = KindInt
	case "int8":
		*k = KindInt8
	case "int16":
		*k = KindInt16
	case "int32":
		*k = KindInt32
	case "int64":
		*k = KindInt64
	case "float32":
		*k = KindFloat32
	case "float64":
		*k = KindFloat64
	case "byte":
		*k = KindByte
	case "bytes":
		*k = KindBytes
	case "string":
		*k = KindString
	case "any":
		*k = KindAny
	case "map":
		*k = KindMap
	case "vector":
		*k = KindVector
	case "array":
		*k = KindArray
	case "enum":
		*k = KindEnum
	case "struct":
		*k = KindStruct
	case "interface":
		*k = KindInterface
	default:
		return fmt.Errorf("invalid kind %q", s)
	}
	return nil
}

// @api(Object/Common/Type/Kind.Bits) returns the number of bits for the type.
// If the type has unknown bits, it returns 0 (for example, `any`, `string`, `bytes`).
func (k Kind) Bits() int {
	switch k {
	case KindInt8, KindByte, KindBool:
		return 8
	case KindInt16:
		return 16
	case KindInt32, KindFloat32, KindInt:
		return 32
	case KindInt64, KindFloat64:
		return 64
	default:
		return 0
	}
}

// @api(Object/Common/Type/Kind.IsPrimitive) reports whether the type is a [PrimitiveType](#Object/PrimitiveType).
func (k Kind) IsPrimitive() bool {
	return k.IsNumeric() || k.IsString() || k.IsBytes() || k.IsBool() || k.IsAny() || k.IsTime() || k.IsDuration()
}

// @api(Object/Common/Type/Kind.IsInteger) reports whether the type is an integer.
// It includes `int`, `int8`, `int16`, `int32`, `int64`, and `byte`.
func (k Kind) IsInteger() bool {
	switch k {
	case KindInt, KindInt8, KindInt16, KindInt32, KindInt64, KindByte:
		return true
	}
	return false
}

// @api(Object/Common/Type/Kind.IsFloat) reports whether the type is a floating point.
// It includes `float32` and `float64`.
func (k Kind) IsFloat() bool {
	switch k {
	case KindFloat32, KindFloat64:
		return true
	}
	return false
}

// @api(Object/Common/Type/Kind.IsNumeric) reports whether the type is a numeric type.
// It includes integer and floating point types.
func (k Kind) IsNumeric() bool { return k.IsInteger() || k.IsFloat() }

// @api(Object/Common/Type/Kind.IsString) reports whether the type is a string.
func (k Kind) IsString() bool { return k == KindString }

// @api(Object/Common/Type/Kind.IsByte) reports whether the type is a byte.
func (k Kind) IsByte() bool { return k == KindByte }

// @api(Object/Common/Type/Kind.IsBytes) reports whether the type is a byte slice.
func (k Kind) IsBytes() bool { return k == KindBytes }

// @api(Object/Common/Type/Kind.IsBool) reports whether the type is a boolean.
func (k Kind) IsBool() bool { return k == KindBool }

// @api(Object/Common/Type/Kind.IsTime) reports whether the type is a time.
func (k Kind) IsTime() bool { return k == KindTime }

// @api(Object/Common/Type/Kind.IsDuration) reports whether the type is a duration.
func (k Kind) IsDuration() bool { return k == KindDuration }

// @api(Object/Common/Type/Kind.IsAny) reports whether the type is any.
func (k Kind) IsAny() bool { return k == KindAny }

// @api(Object/Common/Type/Kind.IsMap) reports whether the type is a map.
func (k Kind) IsMap() bool { return k == KindMap }

// @api(Object/Common/Type/Kind.IsVector) reports whether the type is a vector.
func (k Kind) IsVector() bool { return k == KindVector }

// @api(Object/Common/Type/Kind.IsArray) reports whether the type is an array.
func (k Kind) IsArray() bool { return k == KindArray }

// @api(Object/Common/Type/Kind.IsEnum) reports whether the type is an enumeration.
func (k Kind) IsEnum() bool { return k == KindEnum }

// @api(Object/Common/Type/Kind.IsStruct) reports whether the type is a structure.
func (k Kind) IsStruct() bool { return k == KindStruct }

// @api(Object/Common/Type/Kind.IsInterface) reports whether the type is an interface.
func (k Kind) IsInterface() bool { return k == KindInterface }

// @api(Object/Common/Type/Kind.Compatible) returns the compatible type kind between two kinds.
// If the kinds are not compatible, it returns `KindInvalid`.
// If the kinds are the same, it returns the kind.
// If the kinds are both numeric, it returns the kind with the most bits.
func (k Kind) Compatible(other Kind) Kind {
	if k == other {
		return k
	}
	if k.IsNumeric() && other.IsNumeric() {
		if k.IsFloat() || other.IsFloat() {
			if max(k.Bits(), other.Bits()) == 64 {
				return KindFloat64
			}
			return KindFloat32
		}
		if k.Bits() > other.Bits() {
			return k
		}
		return other
	}
	return KindInvalid
}

var primitiveKinds = []Kind{
	KindBool,
	KindInt,
	KindInt8,
	KindInt16,
	KindInt32,
	KindInt64,
	KindFloat32,
	KindFloat64,
	KindByte,
	KindBytes,
	KindString,
	KindTime,
	KindDuration,
	KindAny,
}

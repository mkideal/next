package types

//go:generate stringer -type=Kind -linecomment
type Kind int

// @api(Object/Common/Type/Kind) represents the type kind.
// Currently, the following kinds are supported:
//
// - `bool`: true or false
// - `int`: integer
// - `int8`: 8-bit integer
// - `int16`: 16-bit integer
// - `int32`: 32-bit integer
// - `int64`: 64-bit integer
// - `float32`: 32-bit floating point
// - `float64`: 64-bit floating point
// - `byte`: byte
// - `bytes`: byte slice
// - `string`: string
// - `any`: any object
// - `map`: dictionary
// - `vector`: vector of elements
// - `array`: array of elements
// - `enum`: enumeration
// - `struct`: structure
// - `interface`: interface
const (
	KindInvalid   Kind = iota // Invalid
	KindBool                  // Bool
	KindInt                   // Int
	KindInt8                  // Int8
	KindInt16                 // Int16
	KindInt32                 // Int32
	KindInt64                 // Int64
	KindFloat32               // Float32
	KindFloat64               // Float64
	KindByte                  // Byte
	KindBytes                 // Bytes
	KindString                // String
	KindAny                   // Any
	KindMap                   // Map
	KindVector                // Vector
	KindArray                 // Array
	KindEnum                  // Enum
	KindStruct                // Struct
	KindInterface             // Interface

	kindCount // -count-
)

// @api(Object/Common/Type/Kind.Valid) represents the type kind.
func (k Kind) Valid() bool {
	return k > KindInvalid && k < kindCount
}

// @api(Object/Common/Type/Kind.Bits) returns the number of bits for the type.
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

// @api(Object/Common/Type/Kind.IsInteger) reports whether the type is an integer.
func (k Kind) IsInteger() bool {
	switch k {
	case KindInt, KindInt8, KindInt16, KindInt32, KindInt64, KindByte:
		return true
	}
	return false
}

// @api(Object/Common/Type/Kind.IsFloat) reports whether the type is a floating point.
func (k Kind) IsFloat() bool {
	switch k {
	case KindFloat32, KindFloat64:
		return true
	}
	return false
}

// @api(Object/Common/Type/Kind.IsNumeric) reports whether the type is a numeric type.
func (k Kind) IsNumeric() bool {
	return k.IsInteger() || k.IsFloat()
}

// @api(Object/Common/Type/Kind.IsString) reports whether the type is a string.
func (k Kind) IsString() bool {
	return k == KindString
}

// @api(Object/Common/Type/Kind.Compatible) returns the compatible type between two types.
// If the types are not compatible, it returns `KindInvalid`.
// If the types are the same, it returns the type.
// If the types are numeric, it returns the type with the most bits.
func (k Kind) Compatible(other Kind) Kind {
	if k == other {
		return k
	}
	if k.IsNumeric() && other.IsNumeric() {
		if k.IsFloat() || other.IsFloat() {
			if k == KindFloat64 || other == KindFloat64 {
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
	KindAny,
}

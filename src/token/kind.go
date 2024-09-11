package token

//go:generate stringer -type=Kind -linecomment
type Kind int

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
)

var PrimitiveKinds = []Kind{
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
